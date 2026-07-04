package randomstring

import (
	"strings"
	"testing"
)

func TestGenerateLength(t *testing.T) {
	for _, length := range []int{0, 1, 7, 32, 100} {
		s, err := Generate(length, "alphanumeric")
		if err != nil {
			t.Fatalf("Generate(%d): %v", length, err)
		}
		if len([]rune(s)) != length {
			t.Errorf("Generate(%d) length = %d, want %d", length, len([]rune(s)), length)
		}
	}
}

func TestNumericOnlyDigits(t *testing.T) {
	s, err := Generate(200, "numeric")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			t.Fatalf("numeric produced non-digit %q in %q", r, s)
		}
	}
}

func TestHexCharset(t *testing.T) {
	s, err := Generate(200, "hex")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range s {
		if !strings.ContainsRune("0123456789abcdef", r) {
			t.Fatalf("hex produced out-of-range char %q in %q", r, s)
		}
	}
}

func TestAlphabeticOnlyLetters(t *testing.T) {
	s, err := Generate(200, "alphabetic")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			t.Fatalf("alphabetic produced non-letter %q in %q", r, s)
		}
	}
}

func TestCustomCharset(t *testing.T) {
	const custom = "XYZ#@!"
	s, err := GenerateFrom(300, custom)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range s {
		if !strings.ContainsRune(custom, r) {
			t.Fatalf("custom charset produced out-of-range char %q", r)
		}
	}
}

func TestSingleCharCharsetRepeats(t *testing.T) {
	s, err := GenerateFrom(10, "q")
	if err != nil {
		t.Fatal(err)
	}
	if s != "qqqqqqqqqq" {
		t.Fatalf("single-char charset = %q, want all q", s)
	}
}

func TestNewIs32Alphanumeric(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 32 {
		t.Fatalf("New() length = %d, want 32", len(s))
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			t.Fatalf("New() produced non-alphanumeric %q", r)
		}
	}
}

func TestUnknownCharset(t *testing.T) {
	if _, err := Generate(5, "nope"); err == nil {
		t.Fatal("expected error for unknown charset")
	}
}
