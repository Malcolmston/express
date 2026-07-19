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
	// zone holds the IPv6 zone identifier (RFC 4007), the text after a "%"
	// in the input, without the "%". Empty for IPv4 and for zoneless IPv6.
	zone string
}

// Parse parses s as an IPv4 or IPv6 address and returns an *Address.
// It returns an error if s is not a valid IP address.
//
// An IPv6 address may carry a zone identifier ("%zone" suffix, RFC 4007); the
// zone is retained and reproduced by String, matching ipaddr.js which keeps the
// zoneIndex through toString. The zone must be non-empty and is only accepted on
// IPv6 addresses.
func Parse(s string) (*Address, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("ipaddr: invalid IP address: " + s)
	}

	base := s
	zone := ""
	if i := strings.IndexByte(s, '%'); i >= 0 {
		base = s[:i]
		zone = s[i+1:]
		// A zone identifier is only meaningful for IPv6 and must be
		// non-empty (ipaddr.js rejects "fe80::%").
		if zone == "" || !strings.Contains(base, ":") {
			return nil, errors.New("ipaddr: invalid IP address: " + s)
		}
	}

	ip := net.ParseIP(base)
	if ip == nil {
		return nil, errors.New("ipaddr: invalid IP address: " + s)
	}

	kind := "ipv6"
	// An address is IPv4 if it has a 4-byte representation and its textual
	// form does not contain a colon (to exclude IPv4-mapped IPv6 literals).
	if ip.To4() != nil && !strings.Contains(base, ":") {
		kind = "ipv4"
	}

	return &Address{ip: ip, kind: kind, orig: base, zone: zone}, nil
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

// String returns the canonical textual representation of the address. For a
// zoned IPv6 address the "%zone" suffix is appended, mirroring ipaddr.js.
func (a *Address) String() string {
	var s string
	if a.kind == "ipv4" {
		if v4 := a.ip.To4(); v4 != nil {
			s = v4.String()
		} else {
			s = a.ip.String()
		}
	} else {
		s = a.ip.String()
	}
	if a.zone != "" {
		s += "%" + a.zone
	}
	return s
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

// specialRange is one entry of a special-range table: a network prefix
// (network) covering the leading bits (bits) is labelled name.
type specialRange struct {
	name    string
	network net.IP
	bits    int
}

// The following tables mirror the SpecialRanges maps in ipaddr.js exactly,
// including their order: classification returns the first range that contains
// the address, so more specific ranges that overlap a broader one (for example
// the ORCHID and benchmarking blocks inside the reserved 2001::/23) are listed
// before it. Anything matching no range is "unicast".
//
// Source: lib/ipaddr.js, IPv4.prototype.SpecialRanges and
// IPv6.prototype.SpecialRanges.
var specialRangesV4 = []specialRange{
	{"unspecified", net.IPv4(0, 0, 0, 0), 8},
	{"broadcast", net.IPv4(255, 255, 255, 255), 32},
	{"multicast", net.IPv4(224, 0, 0, 0), 4},
	{"linkLocal", net.IPv4(169, 254, 0, 0), 16},
	{"loopback", net.IPv4(127, 0, 0, 0), 8},
	{"carrierGradeNat", net.IPv4(100, 64, 0, 0), 10},
	{"private", net.IPv4(10, 0, 0, 0), 8},
	{"private", net.IPv4(172, 16, 0, 0), 12},
	{"private", net.IPv4(192, 168, 0, 0), 16},
	{"reserved", net.IPv4(192, 0, 0, 0), 24},
	{"reserved", net.IPv4(192, 0, 2, 0), 24},
	{"reserved", net.IPv4(192, 88, 99, 0), 24},
	{"reserved", net.IPv4(198, 18, 0, 0), 15},
	{"reserved", net.IPv4(198, 51, 100, 0), 24},
	{"reserved", net.IPv4(203, 0, 113, 0), 24},
	{"reserved", net.IPv4(240, 0, 0, 0), 4},
	{"as112", net.IPv4(192, 175, 48, 0), 24},
	{"as112", net.IPv4(192, 31, 196, 0), 24},
	{"amt", net.IPv4(192, 52, 193, 0), 24},
}

var specialRangesV6 = []specialRange{
	{"unspecified", net.ParseIP("::"), 128},
	{"linkLocal", net.ParseIP("fe80::"), 10},
	{"multicast", net.ParseIP("ff00::"), 8},
	{"loopback", net.ParseIP("::1"), 128},
	{"uniqueLocal", net.ParseIP("fc00::"), 7},
	{"ipv4Mapped", net.ParseIP("::ffff:0:0"), 96},
	{"deprecatedSiteLocal", net.ParseIP("fec0::"), 10},
	{"discard", net.ParseIP("100::"), 64},
	{"rfc6145", net.ParseIP("0:0:0:0:ffff::"), 96},
	{"rfc6052", net.ParseIP("64:ff9b::"), 96},
	{"rfc6052", net.ParseIP("64:ff9b:1::"), 48},
	{"6to4", net.ParseIP("2002::"), 16},
	{"teredo", net.ParseIP("2001::"), 32},
	{"benchmarking", net.ParseIP("2001:2::"), 48},
	{"amt", net.ParseIP("2001:3::"), 32},
	{"as112v6", net.ParseIP("2001:4:112::"), 48},
	{"as112v6", net.ParseIP("2620:4f:8000::"), 48},
	{"deprecatedOrchid", net.ParseIP("2001:10::"), 28},
	{"orchid2", net.ParseIP("2001:20::"), 28},
	{"droneRemoteIdProtocolEntityTags", net.ParseIP("2001:30::"), 28},
	{"segmentRouting", net.ParseIP("5f00::"), 16},
	{"reserved", net.ParseIP("2001::"), 23},
	{"reserved", net.ParseIP("2001:db8::"), 32},
	{"reserved", net.ParseIP("3fff::"), 20},
}

// matchPrefix reports whether the leading bits of ip and network are equal.
// ip and network must be byte slices of the same length.
func matchPrefix(ip, network []byte, bits int) bool {
	if len(ip) != len(network) || bits < 0 || bits > len(ip)*8 {
		return false
	}
	full := bits / 8
	for i := 0; i < full; i++ {
		if ip[i] != network[i] {
			return false
		}
	}
	if rem := bits % 8; rem > 0 {
		mask := byte(0xff) << (8 - rem)
		if ip[full]&mask != network[full]&mask {
			return false
		}
	}
	return true
}

// Range classifies the address, returning a category name.
//
// The categories and the CIDR blocks that define them mirror ipaddr.js's
// SpecialRanges tables. For IPv4 they are: "unspecified", "broadcast",
// "multicast", "linkLocal", "loopback", "carrierGradeNat", "private",
// "reserved", "as112", "amt" and the "unicast" fallthrough. For IPv6 they are:
// "unspecified", "linkLocal", "multicast", "loopback", "uniqueLocal",
// "ipv4Mapped", "deprecatedSiteLocal", "discard", "rfc6145", "rfc6052", "6to4",
// "teredo", "benchmarking", "amt", "as112v6", "deprecatedOrchid", "orchid2",
// "droneRemoteIdProtocolEntityTags", "segmentRouting", "reserved" and the
// "unicast" fallthrough.
func (a *Address) Range() string {
	if a.kind == "ipv4" {
		ip := a.ip.To4()
		if ip == nil {
			return "unicast"
		}
		for _, r := range specialRangesV4 {
			if matchPrefix(ip, r.network.To4(), r.bits) {
				return r.name
			}
		}
		return "unicast"
	}

	ip := a.ip.To16()
	if ip == nil {
		return "unicast"
	}
	for _, r := range specialRangesV6 {
		if matchPrefix(ip, r.network.To16(), r.bits) {
			return r.name
		}
	}
	return "unicast"
}
