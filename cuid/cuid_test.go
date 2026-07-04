package cuid

import (
	"testing"
)

func isLetter(b byte) bool { return b >= 'a' && b <= 'z' }

func TestDefaultLength(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	if len(s) != DefaultLength {
		t.Fatalf("length = %d, want %d: %q", len(s), DefaultLength, s)
	}
}

func TestFirstCharLetterAndCharset(t *testing.T) {
	for i := 0; i < 500; i++ {
		s, err := New()
		if err != nil {
			t.Fatalf("New error: %v", err)
		}
		if !isLetter(s[0]) {
			t.Fatalf("first char not a letter: %q", s)
		}
		for j := 0; j < len(s); j++ {
			c := s[j]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z')) {
				t.Fatalf("invalid char %c in %q", c, s)
			}
		}
	}
}

func TestNewLength(t *testing.T) {
	for _, n := range []int{2, 10, 24, 32} {
		s, err := NewLength(n)
		if err != nil {
			t.Fatalf("NewLength(%d) error: %v", n, err)
		}
		if len(s) != n {
			t.Fatalf("NewLength(%d) length = %d", n, len(s))
		}
		if !isLetter(s[0]) {
			t.Fatalf("first char not a letter: %q", s)
		}
	}
}

func TestUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10000; i++ {
		s, err := New()
		if err != nil {
			t.Fatalf("New error: %v", err)
		}
		if seen[s] {
			t.Fatalf("duplicate id: %q", s)
		}
		seen[s] = true
	}
}

func TestIsCuid(t *testing.T) {
	for i := 0; i < 100; i++ {
		s, _ := New()
		if !IsCuid(s) {
			t.Fatalf("IsCuid(%q) = false, want true", s)
		}
	}
	invalid := []string{
		"",
		"ABC123",                               // uppercase
		"1abc",                                 // starts with digit
		"a",                                    // too short
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // too long (36)
		"abc-123",                              // invalid char
	}
	for _, v := range invalid {
		if IsCuid(v) {
			t.Fatalf("IsCuid(%q) = true, want false", v)
		}
	}
}
