// Package ipaddr parity tests.
//
// The input -> expected-output vectors below are transcribed directly from the
// upstream ipaddr.js test suite and library, not invented, so this port can be
// checked for behavioural parity with the original.
//
// Upstream sources (whitequark/ipaddr.js):
//   Tests: https://raw.githubusercontent.com/whitequark/ipaddr.js/main/test/ipaddr.test.js
//   Lib:   https://raw.githubusercontent.com/whitequark/ipaddr.js/main/lib/ipaddr.js
//
// The Go port implements a subset of ipaddr.js (Parse/IsValid, Kind, String,
// Match(cidr), Range). Vectors that exercise features the port does not model
// (the fromByteArray / octet / parts constructors, parseCIDR objects,
// prefixLengthFromSubnetMask, subnetMask/broadcast/network derivation, the
// numeric/octal/hex "weird format" IPv4 parser, and ipaddr.process's narrowing
// of IPv4-mapped IPv6 to ipv4) are intentionally not encoded here; those gaps
// are recorded in the task notes rather than tested.

package ipaddr

import "testing"

// TestParityKind mirrors upstream "is able to determine IP address type".
func TestParityKind(t *testing.T) {
	cases := map[string]string{
		"8.8.8.8":            "ipv4",
		"2001:db8:3312::1":   "ipv6",
		"2001:db8:3312::1%z": "ipv6", // zoneIndex is retained, kind stays ipv6
	}
	for in, want := range cases {
		a, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): unexpected error %v", in, err)
		}
		if got := a.Kind(); got != want {
			t.Errorf("Kind(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestParityIsValid mirrors upstream isValid vectors, restricted to the textual
// forms net.ParseIP (and hence this port) accepts. The numeric/octal/hex IPv4
// forms that ipaddr.js additionally accepts are recorded as gaps, not tested.
func TestParityIsValid(t *testing.T) {
	cases := map[string]bool{
		"2001:db8:F53A::1":     true,
		"::ffff:192.168.1.1":   true,
		"::ffff:192.168.1.1%z": true, // zoned IPv6 is valid
		"::1.2.3.4%z":          true,
		"::%z":                 true,
		"2001:db8::F53A::1":    false, // two "::" compressions
		"fe80::wtf":            false,
		"fe80::%":              false, // empty zone
		"2002::2:":             false,
		"::ffff:300.168.1.1":   false, // octet > 255
		"4999999999":           false, // too large to be an address
		"-1":                   false,
	}
	for in, want := range cases {
		if got := IsValid(in); got != want {
			t.Errorf("IsValid(%q) = %v, want %v", in, got, want)
		}
	}
}

// TestParityString mirrors upstream toString/zoneIndex vectors, restricted to
// forms whose canonical rendering agrees between net.IP.String and ipaddr.js
// (i.e. excluding IPv4-mapped addresses, which the stdlib prints dotted and
// ipaddr.js prints as hextets).
func TestParityString(t *testing.T) {
	cases := map[string]string{
		"2001:db8:f53a::1%2":   "2001:db8:f53a::1%2",
		"2001:db8:f53a::1%WAT": "2001:db8:f53a::1%WAT",
		"2001:db8:f53a::1%sUp": "2001:db8:f53a::1%sUp",
		"fe80::%eth0":          "fe80::%eth0",
		"::%eth0":              "::%eth0",
	}
	for in, want := range cases {
		a, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): unexpected error %v", in, err)
		}
		if got := a.String(); got != want {
			t.Errorf("String(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestParityRangeV4 mirrors upstream "detects reserved IPv4 networks".
func TestParityRangeV4(t *testing.T) {
	cases := map[string]string{
		"0.0.0.0":         "unspecified",
		"0.1.0.0":         "unspecified", // 0.0.0.0/8, not just the all-zero host
		"10.1.0.1":        "private",
		"100.64.0.0":      "carrierGradeNat",
		"100.127.255.255": "carrierGradeNat",
		"192.52.193.1":    "amt",
		"192.168.2.1":     "private",
		"192.175.48.0":    "as112",
		"224.100.0.1":     "multicast",
		"169.254.15.0":    "linkLocal",
		"127.1.1.1":       "loopback",
		"255.255.255.255": "broadcast",
		"240.1.2.3":       "reserved",
		"8.8.8.8":         "unicast",
	}
	for in, want := range cases {
		a, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): unexpected error %v", in, err)
		}
		if got := a.Range(); got != want {
			t.Errorf("Range(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestParityRangeV6 mirrors upstream "detects reserved IPv6 networks".
func TestParityRangeV6(t *testing.T) {
	cases := map[string]string{
		"::":                        "unspecified",
		"fe80::1234:5678:abcd:0123": "linkLocal",
		"ff00::1234":                "multicast",
		"::1":                       "loopback",
		"100::42":                   "discard",
		"fc00::":                    "uniqueLocal",
		"::ffff:192.168.1.10":       "ipv4Mapped",
		"fec0::1234":                "deprecatedSiteLocal",
		"::ffff:0:192.168.1.10":     "rfc6145",
		"64:ff9b::1234":             "rfc6052",
		"64:ff9b:1::1234":           "rfc6052",
		"2002:1f63:45e8::1":         "6to4",
		"2001::4242":                "teredo",
		"2001:2::":                  "benchmarking",
		"2001:3::":                  "amt",
		"2001:4:112::":              "as112v6",
		"2620:4f:8000::":            "as112v6",
		"2001:10::":                 "deprecatedOrchid",
		"2001:20::":                 "orchid2",
		"2001:30::":                 "droneRemoteIdProtocolEntityTags",
		"2001:db8::3210":            "reserved",
		"2001:470:8:66::1":          "unicast",
		"2001:470:8:66::1%z":        "unicast",
		"5f00::1":                   "segmentRouting",
		"3fff::1":                   "reserved",
	}
	for in, want := range cases {
		a, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): unexpected error %v", in, err)
		}
		if got := a.Range(); got != want {
			t.Errorf("Range(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestParityMatchV4 mirrors upstream "matches IPv4 CIDR correctly" for the
// vectors expressible as a CIDR-string containment check.
func TestParityMatchV4(t *testing.T) {
	// addr 10.5.0.1 against (cidr -> expected containment)
	cases := map[string]bool{
		"0.0.0.0/0":   true,
		"11.0.0.0/8":  false,
		"10.0.0.0/8":  true,
		"10.5.5.0/16": true,
		"10.4.5.0/16": false,
		"10.4.5.0/15": true,
		"10.5.0.2/32": false,
		"10.5.0.1/32": true,
	}
	a, err := Parse("10.5.0.1")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	for cidr, want := range cases {
		got, err := a.Match(cidr)
		if err != nil {
			t.Fatalf("Match(%q): unexpected error %v", cidr, err)
		}
		if got != want {
			t.Errorf("Match(10.5.0.1, %q) = %v, want %v", cidr, got, want)
		}
	}
}

// TestParityMatchV6 mirrors upstream "matches IPv6 CIDR correctly" for the
// vectors expressible as a CIDR-string containment check.
func TestParityMatchV6(t *testing.T) {
	// addr 2001:db8:f53a::1 against (cidr -> expected containment)
	cases := map[string]bool{
		"::/0":                  true,
		"2001:db8:f53a::1:1/64": true,
		"2001:db8:f53b::1:1/48": false,
		"2001:db8:f531::1:1/44": true,
		"2001:db8:f500::1/40":   true,
		"2001:db9:f500::1/40":   false,
		"2001:db8:f53a::1/128":  true,
	}
	a, err := Parse("2001:db8:f53a::1")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	for cidr, want := range cases {
		got, err := a.Match(cidr)
		if err != nil {
			t.Fatalf("Match(%q): unexpected error %v", cidr, err)
		}
		if got != want {
			t.Errorf("Match(2001:db8:f53a::1, %q) = %v, want %v", cidr, got, want)
		}
	}
}
