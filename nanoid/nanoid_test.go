package nanoid

import (
	"strings"
	"testing"
)

func TestDefaultLength(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	if len(s) != DefaultSize {
		t.Fatalf("length = %d, want %d", len(s), DefaultSize)
	}
	for _, c := range s {
		if !strings.ContainsRune(DefaultAlphabet, c) {
			t.Fatalf("char %c not in default alphabet", c)
		}
	}
}

func TestSizeHonored(t *testing.T) {
	for _, n := range []int{1, 5, 10, 32, 128} {
		s, err := NewSize(n)
		if err != nil {
			t.Fatalf("NewSize(%d) error: %v", n, err)
		}
		if len(s) != n {
			t.Fatalf("NewSize(%d) length = %d", n, len(s))
		}
	}
}

func TestCustomAlphabet(t *testing.T) {
	alphabet := "abcdef"
	for i := 0; i < 50; i++ {
		s, err := Custom(alphabet, 30)
		if err != nil {
			t.Fatalf("Custom error: %v", err)
		}
		if len(s) != 30 {
			t.Fatalf("length = %d, want 30", len(s))
		}
		for _, c := range s {
			if !strings.ContainsRune(alphabet, c) {
				t.Fatalf("char %c not in custom alphabet", c)
			}
		}
	}
}

func TestSingleCharAlphabet(t *testing.T) {
	s, err := Custom("x", 10)
	if err != nil {
		t.Fatalf("Custom error: %v", err)
	}
	if s != "xxxxxxxxxx" {
		t.Fatalf("single char = %q, want xxxxxxxxxx", s)
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

func TestInvalid(t *testing.T) {
	if _, err := NewSize(0); err == nil {
		t.Fatal("expected error for size 0")
	}
	if _, err := Custom("", 5); err == nil {
		t.Fatal("expected error for empty alphabet")
	}
}
