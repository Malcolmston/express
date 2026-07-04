// Package ipaddr provides parsing and classification of IPv4 and IPv6
// addresses. It is a small Go port of a subset of the JavaScript "ipaddr.js"
// library, built on top of the standard library net package.
package ipaddr

import (
	"errors"
	"net"
	"strings"
)

// Address represents a parsed IPv4 or IPv6 address.
type Address struct {
	ip   net.IP
	kind string // "ipv4" or "ipv6"
	// orig preserves the original textual form so String reflects the input
	// family (an IPv4 string stays IPv4 rather than being widened).
	orig string
}

// Parse parses s as an IPv4 or IPv6 address and returns an *Address.
// It returns an error if s is not a valid IP address.
func Parse(s string) (*Address, error) {
	s = strings.TrimSpace(s)
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, errors.New("ipaddr: invalid IP address: " + s)
	}

	kind := "ipv6"
	// An address is IPv4 if it has a 4-byte representation and its textual
	// form does not contain a colon (to exclude IPv4-mapped IPv6 literals).
	if ip.To4() != nil && !strings.Contains(s, ":") {
		kind = "ipv4"
	}

	return &Address{ip: ip, kind: kind, orig: s}, nil
}

// IsValid reports whether s is a valid IPv4 or IPv6 address.
func IsValid(s string) bool {
	_, err := Parse(s)
	return err == nil
}

// Kind returns the address family: "ipv4" or "ipv6".
func (a *Address) Kind() string {
	return a.kind
}

// String returns the canonical textual representation of the address.
func (a *Address) String() string {
	if a.kind == "ipv4" {
		if v4 := a.ip.To4(); v4 != nil {
			return v4.String()
		}
	}
	return a.ip.String()
}

// Match reports whether the address is contained within the given CIDR range,
// for example "192.168.0.0/16" or "2001:db8::/32". It returns an error if the
// CIDR is invalid.
func (a *Address) Match(cidr string) (bool, error) {
	_, network, err := net.ParseCIDR(strings.TrimSpace(cidr))
	if err != nil {
		return false, err
	}
	return network.Contains(a.ip), nil
}

// Range classifies the address, returning a category name.
//
// For IPv4 addresses the categories are: "unspecified", "broadcast",
// "multicast", "linkLocal", "loopback", "private", "reserved", and
// "unicast" (the default for ordinary public addresses).
//
// For IPv6 addresses the categories are: "unspecified", "loopback",
// "multicast", "linkLocal", "uniqueLocal", "ipv4Mapped", and "unicast".
func (a *Address) Range() string {
	if a.kind == "ipv4" {
		return a.rangeV4()
	}
	return a.rangeV6()
}

func (a *Address) rangeV4() string {
	ip := a.ip.To4()
	if ip == nil {
		return "unicast"
	}

	switch {
	case ip.Equal(net.IPv4zero):
		return "unspecified"
	case ip[0] == 255 && ip[1] == 255 && ip[2] == 255 && ip[3] == 255:
		return "broadcast"
	case ip[0] >= 224 && ip[0] <= 239:
		return "multicast"
	case ip[0] == 169 && ip[1] == 254:
		return "linkLocal"
	case ip[0] == 127:
		return "loopback"
	case ip[0] == 10,
		ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31,
		ip[0] == 192 && ip[1] == 168:
		return "private"
	case ip[0] == 192 && ip[1] == 0 && ip[2] == 0,
		ip[0] == 192 && ip[1] == 0 && ip[2] == 2,
		ip[0] == 192 && ip[1] == 88 && ip[2] == 99,
		ip[0] == 198 && (ip[1] == 18 || ip[1] == 19),
		ip[0] == 198 && ip[1] == 51 && ip[2] == 100,
		ip[0] == 203 && ip[1] == 0 && ip[2] == 113,
		ip[0] >= 240:
		return "reserved"
	default:
		return "unicast"
	}
}

func (a *Address) rangeV6() string {
	ip := a.ip.To16()
	if ip == nil {
		return "unicast"
	}

	switch {
	case ip.Equal(net.IPv6unspecified):
		return "unspecified"
	case ip.Equal(net.IPv6loopback):
		return "loopback"
	case ip[0] == 0xff:
		return "multicast"
	case ip[0] == 0xfe && (ip[1]&0xc0) == 0x80:
		return "linkLocal"
	case (ip[0] & 0xfe) == 0xfc:
		return "uniqueLocal"
	case a.ip.To4() != nil:
		return "ipv4Mapped"
	default:
		return "unicast"
	}
}
