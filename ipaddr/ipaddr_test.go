package ipaddr

import "testing"

func TestParseIPv4(t *testing.T) {
	a, err := Parse("192.168.1.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Kind() != "ipv4" {
		t.Errorf("Kind = %q, want ipv4", a.Kind())
	}
	if a.String() != "192.168.1.1" {
		t.Errorf("String = %q, want 192.168.1.1", a.String())
	}
}

func TestParseIPv6(t *testing.T) {
	a, err := Parse("2001:db8::1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Kind() != "ipv6" {
		t.Errorf("Kind = %q, want ipv6", a.Kind())
	}
	if a.String() != "2001:db8::1" {
		t.Errorf("String = %q, want 2001:db8::1", a.String())
	}
}

func TestParseInvalid(t *testing.T) {
	for _, s := range []string{"", "not-an-ip", "999.999.999.999", "192.168.1"} {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) expected error", s)
		}
	}
}

func TestIsValid(t *testing.T) {
	cases := map[string]bool{
		"127.0.0.1":  true,
		"::1":        true,
		"10.0.0.256": false,
		"garbage":    false,
	}
	for s, want := range cases {
		if got := IsValid(s); got != want {
			t.Errorf("IsValid(%q) = %v, want %v", s, got, want)
		}
	}
}

func TestMatch(t *testing.T) {
	a, _ := Parse("192.168.5.5")
	ok, err := a.Match("192.168.0.0/16")
	if err != nil || !ok {
		t.Errorf("Match in range: ok=%v err=%v", ok, err)
	}
	ok, err = a.Match("10.0.0.0/8")
	if err != nil || ok {
		t.Errorf("Match out of range: ok=%v err=%v", ok, err)
	}
}

func TestMatchIPv6(t *testing.T) {
	a, _ := Parse("2001:db8::5")
	ok, err := a.Match("2001:db8::/32")
	if err != nil || !ok {
		t.Errorf("Match v6 in range: ok=%v err=%v", ok, err)
	}
}

func TestMatchInvalidCIDR(t *testing.T) {
	a, _ := Parse("192.168.1.1")
	if _, err := a.Match("not-a-cidr"); err == nil {
		t.Error("expected error for invalid CIDR")
	}
}

func TestRangeV4(t *testing.T) {
	cases := map[string]string{
		"0.0.0.0":         "unspecified",
		"255.255.255.255": "broadcast",
		"224.0.0.1":       "multicast",
		"169.254.1.1":     "linkLocal",
		"127.0.0.1":       "loopback",
		"10.1.2.3":        "private",
		"172.16.5.5":      "private",
		"192.168.1.1":     "private",
		"8.8.8.8":         "unicast",
		"192.0.2.1":       "reserved",
		"240.0.0.1":       "reserved",
	}
	for s, want := range cases {
		a, err := Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q): %v", s, err)
		}
		if got := a.Range(); got != want {
			t.Errorf("Range(%q) = %q, want %q", s, got, want)
		}
	}
}

func TestRangeV6(t *testing.T) {
	cases := map[string]string{
		"::":      "unspecified",
		"::1":     "loopback",
		"ff02::1": "multicast",
		"fe80::1": "linkLocal",
		"fc00::1": "uniqueLocal",
		"fd00::1": "uniqueLocal",
		// 2001:db8::/32 is the RFC 3849 documentation block, which ipaddr.js
		// classifies as "reserved".
		"2001:db8::1": "reserved",
	}
	for s, want := range cases {
		a, err := Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q): %v", s, err)
		}
		if got := a.Range(); got != want {
			t.Errorf("Range(%q) = %q, want %q", s, got, want)
		}
	}
}

func TestParseTrimsWhitespace(t *testing.T) {
	a, err := Parse("  10.0.0.1  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.String() != "10.0.0.1" {
		t.Errorf("String = %q", a.String())
	}
}
