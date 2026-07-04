package uuid

import (
	"strings"
	"testing"
)

func TestV4Format(t *testing.T) {
	for i := 0; i < 100; i++ {
		s, err := V4()
		if err != nil {
			t.Fatalf("V4 error: %v", err)
		}
		if len(s) != 36 {
			t.Fatalf("V4 length = %d, want 36: %q", len(s), s)
		}
		if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
			t.Fatalf("V4 dashes wrong: %q", s)
		}
		if s[14] != '4' {
			t.Fatalf("V4 version nibble = %c, want 4: %q", s[14], s)
		}
		if !strings.ContainsRune("89ab", rune(s[19])) {
			t.Fatalf("V4 variant char = %c, want [89ab]: %q", s[19], s)
		}
		if !Validate(s) {
			t.Fatalf("V4 output not valid: %q", s)
		}
	}
}

func TestV5Known(t *testing.T) {
	got, err := V5(NamespaceDNS, "example.com")
	if err != nil {
		t.Fatalf("V5 error: %v", err)
	}
	want := "cfbff0d1-9375-5685-968c-48ce8b15ae17"
	if got != want {
		t.Fatalf("V5 = %q, want %q", got, want)
	}
	if got[14] != '5' {
		t.Fatalf("V5 version nibble = %c, want 5", got[14])
	}
}

func TestValidate(t *testing.T) {
	valid := []string{
		"cfbff0d1-9375-5685-968c-48ce8b15ae17",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}
	for _, v := range valid {
		if !Validate(v) {
			t.Fatalf("Validate(%q) = false, want true", v)
		}
	}
	invalid := []string{
		"",
		"not-a-uuid",
		"cfbff0d1-9375-5685-968c-48ce8b15ae1",  // too short
		"cfbff0d19375-5685-968c-48ce8b15ae17",  // missing dash
		"zfbff0d1-9375-5685-968c-48ce8b15ae17", // bad hex
	}
	for _, v := range invalid {
		if Validate(v) {
			t.Fatalf("Validate(%q) = true, want false", v)
		}
	}
}

func TestParseFormatRoundTrip(t *testing.T) {
	s := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	b, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if Format(b) != s {
		t.Fatalf("round trip = %q, want %q", Format(b), s)
	}
}
