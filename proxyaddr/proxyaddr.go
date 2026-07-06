// Package proxyaddr determines the real client address of a request that has
// passed through one or more trusted reverse proxies. It is a Go port of the
// npm module "proxy-addr", the same logic Express uses to implement its
// "trust proxy" setting and the req.ip / req.ips accessors.
//
// When an application sits behind a load balancer, CDN, or reverse proxy, the
// socket's remote address is the address of the nearest proxy rather than the
// end user. Each hop is expected to append the address it saw to the
// X-Forwarded-For header, producing a left-to-right list that runs from the
// original client through every intermediary. Because any of those values can
// be forged by a client, they can only be trusted up to the point where the
// chain leaves infrastructure you control. This package encodes that decision:
// given the socket address, the header, and a predicate describing which hops
// are trusted, it returns the closest address that is not itself a trusted
// proxy.
//
// The forwarded chain is derived from the socket remote address plus the
// X-Forwarded-For header, in the same reverse ordering used by the "forwarded"
// module: the socket address comes first, followed by the header entries read
// from rightmost (nearest proxy) to leftmost (original client). Starting from
// the socket address, the chain is walked while each address is trusted; the
// first untrusted address is the real client address. If every address in the
// chain is trusted, the topmost (leftmost header) address is returned. All
// returns that full ordered chain, ProxyAddr returns the single resolved
// client address, and All combined with a trust predicate lets callers build
// the equivalent of Express's req.ips.
//
// The trust predicate is a func(addr string, i int) bool that receives each
// candidate address and its index in the chain. Callers may supply arbitrary
// logic, but Compile builds a predicate from a list of trusted CIDR strings,
// bare IP addresses (treated as single-host /32 or /128 ranges), and the named
// presets recognized by proxy-addr. The presets are "loopback"
// (127.0.0.1/8 and ::1/128), "linklocal" (169.254.0.0/16 and fe80::/10), and
// "uniquelocal" (the RFC 1918 private IPv4 blocks 10.0.0.0/8, 172.16.0.0/12,
// 192.168.0.0/16 plus the IPv6 unique-local block fc00::/7). Preset names are
// matched case-insensitively.
//
// Edge cases follow the Node semantics closely. Addresses may carry a port
// (including bracketed IPv6 literals such as "[::1]:8080"); the port is
// stripped before comparison. An empty X-Forwarded-For yields a chain of just
// the socket address. Compile returns an error for an empty trust value or any
// entry that is neither a preset nor a valid CIDR/IP, and the predicate it
// returns reports false for addresses that fail to parse rather than treating
// them as trusted. The chief parity gap versus the JavaScript original is that
// this port operates on the raw remote-address and header strings a Go handler
// already has, rather than on a Node request object.
package proxyaddr

import (
	"errors"
	"net"
	"strings"
)

// presets maps the named CIDR presets to their address ranges.
var presets = map[string][]string{
	"loopback":    {"127.0.0.1/8", "::1/128"},
	"linklocal":   {"169.254.0.0/16", "fe80::/10"},
	"uniquelocal": {"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "fc00::/7"},
}

// forwarded returns the address chain for the given socket remote address and
// X-Forwarded-For header value, matching the "forwarded" module ordering.
func forwarded(remoteAddr, xff string) []string {
	addrs := []string{stripPort(remoteAddr)}
	if xff == "" {
		return addrs
	}
	parts := strings.Split(xff, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		addrs = append(addrs, strings.TrimSpace(parts[i]))
	}
	return addrs
}

// All returns the full forwarded address chain for the request, starting with
// the socket remote address followed by the X-Forwarded-For entries in reverse
// order.
func All(remoteAddr, xff string) []string {
	return forwarded(remoteAddr, xff)
}

// ProxyAddr walks the forwarded chain and returns the first address that is not
// trusted according to trust. The trust function receives each candidate
// address and its index in the chain. Walking starts at the socket address and
// continues up the chain while the current address is trusted. If every address
// in the chain is trusted, the topmost (leftmost header) address is returned.
func ProxyAddr(remoteAddr, xff string, trust func(addr string, i int) bool) string {
	addrs := forwarded(remoteAddr, xff)

	result := addrs[0]
	for i := 0; i < len(addrs); i++ {
		result = addrs[i]
		if i+1 >= len(addrs) {
			break
		}
		if !trust(addrs[i], i) {
			break
		}
	}
	return result
}

// Compile builds a trust function from a list of trusted CIDR strings and/or
// the named presets "loopback", "linklocal", and "uniquelocal". The returned
// function reports whether a given address falls within any of the trusted
// ranges. It returns an error if any value is not a valid preset or CIDR.
func Compile(vals []string) (func(addr string, i int) bool, error) {
	var nets []*net.IPNet

	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v == "" {
			return nil, errors.New("proxyaddr: empty trust value")
		}

		if ranges, ok := presets[strings.ToLower(v)]; ok {
			for _, r := range ranges {
				n, err := parseRange(r)
				if err != nil {
					return nil, err
				}
				nets = append(nets, n)
			}
			continue
		}

		n, err := parseRange(v)
		if err != nil {
			return nil, err
		}
		nets = append(nets, n)
	}

	return func(addr string, _ int) bool {
		ip := net.ParseIP(stripPort(addr))
		if ip == nil {
			return false
		}
		for _, n := range nets {
			if n.Contains(ip) {
				return true
			}
		}
		return false
	}, nil
}

// parseRange parses a CIDR string, or a bare IP address which is treated as a
// single-host range.
func parseRange(s string) (*net.IPNet, error) {
	if strings.Contains(s, "/") {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return nil, err
		}
		return n, nil
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return nil, errors.New("proxyaddr: invalid IP address: " + s)
	}
	bits := 32
	if ip.To4() == nil {
		bits = 128
	}
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)}, nil
}

// stripPort removes a trailing port from an address if present, handling IPv6
// literals (bracketed or bare) as well as IPv4 addresses.
func stripPort(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return addr
	}
	if addr[0] == '[' {
		if host, _, err := net.SplitHostPort(addr); err == nil {
			return host
		}
		if end := strings.IndexByte(addr, ']'); end != -1 {
			return addr[1:end]
		}
		return addr
	}
	if net.ParseIP(addr) != nil {
		return addr
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}
