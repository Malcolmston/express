// Package ipaddr provides parsing and classification of IPv4 and IPv6
// addresses. It is a small Go port of a subset of the JavaScript "ipaddr.js"
// library, built entirely on top of the standard library net package so that no
// third-party dependency is required.
//
// The ipaddr.js module is widely used in the Node ecosystem (notably by Express
// and proxy middleware) to answer questions such as "is this a valid address?",
// "which family is it?", "does it fall inside this CIDR block?" and "what kind
// of address is it — loopback, private, multicast, public unicast?". This port
// exists to answer the same questions in Go programs that need to reason about
// client addresses, trust proxies, or filter traffic, while keeping the small,
// object-oriented feel of the original API: Parse yields an *Address value whose
// methods (Kind, String, Match, Range) mirror the JavaScript instance methods.
//
// Parsing delegates to net.ParseIP after trimming surrounding whitespace, so any
// textual form the standard library accepts is accepted here. The address family
// is decided by inspecting the parsed value: an address that has a 4-byte
// representation and whose input text contains no colon is classified as "ipv4",
// and everything else as "ipv6". That colon check is deliberate — it keeps
// IPv4-mapped IPv6 literals such as "::ffff:1.2.3.4" on the IPv6 side rather than
// silently narrowing them, and the original input string is retained so that
// String reflects the family the caller actually wrote. Match parses a CIDR with
// net.ParseCIDR and reports whether the address is contained in that network,
// which works uniformly for both IPv4 and IPv6 ranges.
//
// Range classifies an address into a category name closely following
// ipaddr.js's special-range tables. For IPv4 the categories are "unspecified",
// "broadcast", "multicast", "linkLocal", "loopback", "private", "reserved" and
// "unicast", where "unicast" is the fallthrough for ordinary public addresses.
// For IPv6 the categories are "unspecified", "loopback", "multicast",
// "linkLocal", "uniqueLocal", "ipv4Mapped" and "unicast". The checks are applied
// in order and the first match wins, so, for example, the all-zeros address is
// reported as "unspecified" before the more general unicast fallthrough is
// reached.
//
// Edge cases are handled explicitly. Parse returns an error for empty input and
// for anything net.ParseIP rejects (for example "999.999.999.999" or a bare
// "192.168.1"); IsValid is a thin boolean wrapper over Parse for callers that
// only need a yes/no answer. Match returns a non-nil error when the CIDR string
// is malformed rather than reporting a false negative. Compared with Node, this
// port intentionally covers only address parsing, validity, family detection,
// CIDR containment and range classification; it does not implement ipaddr.js
// features such as address arithmetic, subnet-mask/prefix conversion, the
// fromByteArray constructors, or the full set of narrowly scoped reserved
// sub-ranges, and its "unicast" bucket therefore absorbs a few blocks that
// ipaddr.js would name individually.
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
