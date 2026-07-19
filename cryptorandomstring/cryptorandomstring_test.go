package cryptorandomstring

import (
	"strings"
	"testing"
)

func TestHexLengthAndCharset(t *testing.T) {
	for _, length := range []int{0, 1, 8, 16, 64} {
		s, err := Hex(length)
		if err != nil {
			t.Fatalf("Hex(%d): %v", length, err)
		}
		if len(s) != length {
			t.Errorf("Hex(%d) length = %d, want %d", length, len(s), length)
		}
		for _, r := range s {
			if !strings.ContainsRune("0123456789abcdef", r) {
				t.Fatalf("hex out-of-range char %q", r)
			}
		}
	}
}

func TestNumericOnlyDigits(t *testing.T) {
	s, err := Generate(Options{Length: 200, Type: "numeric"})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			t.Fatalf("numeric produced non-digit %q", r)
		}
	}
}

func TestURLSafeCharset(t *testing.T) {
	s, err := Generate(Options{Length: 300, Type: "url-safe"})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range s {
		ok := (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' || r == '~'
		if !ok {
			t.Fatalf("url-safe out-of-range char %q", r)
		}
	}
}

func TestCustomCharacters(t *testing.T) {
	const custom = "abc"
	s, err := Generate(Options{Length: 300, Characters: custom})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range s {
		if !strings.ContainsRune(custom, r) {
			t.Fatalf("custom characters out-of-range char %q", r)
		}
	}
}

func TestCharactersOverridesType(t *testing.T) {
	s, err := Generate(Options{Length: 50, Type: "hex", Characters: "Z"})
	if err != nil {
		t.Fatal(err)
	}
	if s != strings.Repeat("Z", 50) {
		t.Fatalf("Characters did not override Type: %q", s)
	}
}

func TestLengthHonoredAcrossTypes(t *testing.T) {
	types := []string{"hex", "base64", "url-safe", "numeric", "distinguishable", "ascii-printable", "alphanumeric"}
	for _, typ := range types {
		s, err := Generate(Options{Length: 25, Type: typ})
		if err != nil {
			t.Fatalf("type %q: %v", typ, err)
		}
		if len([]rune(s)) != 25 {
			t.Errorf("type %q length = %d, want 25", typ, len([]rune(s)))
		}
	}
}

func TestUnknownType(t *testing.T) {
	if _, err := Generate(Options{Length: 5, Type: "bogus"}); err == nil {
		t.Fatal("expected error for unknown type")
	}
}
