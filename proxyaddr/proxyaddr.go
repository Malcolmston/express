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
	"strconv"
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

// subnet is a trusted address range expressed in a unified 128-bit space. Every
// range is normalized to a 16-byte IP (IPv4 ranges use the ::ffff:0:0/96
// IPv4-mapped prefix) plus a prefix length measured in that 128-bit space, so a
// single prefix comparison handles IPv4, IPv6, and IPv4-mapped-IPv6 addresses
// the same way the upstream ipaddr.js matcher does.
type subnet struct {
	ip   net.IP
	ones int
}

// match reports whether the given 16-byte address falls within the subnet.
func (s subnet) match(ip16 net.IP) bool {
	full := s.ones / 8
	rem := s.ones % 8
	for i := 0; i < full; i++ {
		if ip16[i] != s.ip[i] {
			return false
		}
	}
	if rem > 0 {
		mask := byte(0xff) << uint(8-rem)
		if (ip16[full]^s.ip[full])&mask != 0 {
			return false
		}
	}
	return true
}

// Compile builds a trust function from a list of trusted CIDR strings and/or
// the named presets "loopback", "linklocal", and "uniquelocal". The returned
// function reports whether a given address falls within any of the trusted
// ranges. It returns an error if any value is not a valid preset or CIDR.
//
// Accepted range notations mirror proxy-addr: a bare IP address (a single
// host), an IP with a numeric CIDR prefix, and an IPv4 address with an IPv4
// subnet mask ("10.0.0.0/255.0.0.0"). IPv4-mapped IPv6 ranges such as
// "::ffff:a00:0/120" match plain IPv4 addresses, and vice versa.
func Compile(vals []string) (func(addr string, i int) bool, error) {
	var subnets []subnet

	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v == "" {
			return nil, errors.New("proxyaddr: empty trust value")
		}

		if ranges, ok := presets[strings.ToLower(v)]; ok {
			for _, r := range ranges {
				s, err := parseNotation(r)
				if err != nil {
					return nil, err
				}
				subnets = append(subnets, s)
			}
			continue
		}

		s, err := parseNotation(v)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, s)
	}

	return func(addr string, _ int) bool {
		ip := net.ParseIP(stripPort(addr))
		if ip == nil {
			return false
		}
		ip16 := ip.To16()
		if ip16 == nil {
			return false
		}
		for _, s := range subnets {
			if s.match(ip16) {
				return true
			}
		}
		return false
	}, nil
}

// parseNotation parses one trust range: a bare IP (single host), an IP with a
// numeric CIDR prefix, or an IPv4 address with an IPv4 subnet mask. The result
// is normalized into the unified 128-bit space used by subnet.match.
func parseNotation(note string) (subnet, error) {
	pos := strings.LastIndexByte(note, '/')
	str := note
	if pos != -1 {
		str = note[:pos]
	}

	ip := net.ParseIP(str)
	if ip == nil {
		return subnet{}, errors.New("proxyaddr: invalid IP address: " + str)
	}

	// The IPv6 vs IPv4 "kind" follows the textual form, matching ipaddr.js:
	// "::ffff:a00:2" is treated as IPv6 (max prefix 128) even though it maps to
	// an IPv4 address, while "10.0.0.2" is IPv4 (max prefix 32).
	ipv6Family := strings.Contains(str, ":")
	max := 32
	if ipv6Family {
		max = 128
	}

	n := max
	if pos != -1 {
		rng := note[pos+1:]
		switch {
		case isDigits(rng):
			n, _ = strconv.Atoi(rng)
		case !ipv6Family && net.ParseIP(rng) != nil:
			// IPv4 subnet-mask notation, e.g. "255.255.255.0".
			m4 := net.ParseIP(rng).To4()
			if m4 == nil {
				return subnet{}, errors.New("proxyaddr: invalid range on address: " + note)
			}
			ones, bits := net.IPMask(m4).Size()
			if bits == 0 {
				// Non-contiguous mask.
				return subnet{}, errors.New("proxyaddr: invalid range on address: " + note)
			}
			n = ones
		default:
			return subnet{}, errors.New("proxyaddr: invalid range on address: " + note)
		}
	}

	if n <= 0 || n > max {
		return subnet{}, errors.New("proxyaddr: invalid range on address: " + note)
	}

	ones := n
	if !ipv6Family {
		// Shift an IPv4 prefix into the ::ffff:0:0/96 mapped space.
		ones = 96 + n
	}
	return subnet{ip: ip.To16(), ones: ones}, nil
}

// isDigits reports whether s is a non-empty run of ASCII digits, matching the
// upstream /^[0-9]+$/ test used to distinguish a numeric CIDR prefix.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
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
