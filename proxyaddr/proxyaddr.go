// Package proxyaddr determines the real client address of a request that has
// passed through one or more trusted reverse proxies. It is a Go port of the
// npm module "proxy-addr".
//
// The forwarded chain is derived from the socket remote address plus the
// X-Forwarded-For header, in the same reverse ordering used by the "forwarded"
// module (socket address first, then header entries from rightmost to
// leftmost). Starting from the socket address, the chain is walked while each
// address is trusted; the first untrusted address is the real client address.
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
