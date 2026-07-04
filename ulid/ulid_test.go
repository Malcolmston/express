package ulid

import (
	"testing"
)

func TestLength(t *testing.T) {
	s, err := New(1469918176385)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	if len(s) != 26 {
		t.Fatalf("length = %d, want 26: %q", len(s), s)
	}
}

func TestDeterministic(t *testing.T) {
	entropy := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	a, err := NewWithEntropy(1469918176385, entropy)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	b, err := NewWithEntropy(1469918176385, entropy)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if a != b {
		t.Fatalf("not deterministic: %q vs %q", a, b)
	}
}

func TestTimestampRoundTrip(t *testing.T) {
	entropy := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, ms := range []uint64{0, 1, 1469918176385, (1 << 48) - 1} {
		s, err := NewWithEntropy(ms, entropy)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		got, err := Timestamp(s)
		if err != nil {
			t.Fatalf("Timestamp error: %v", err)
		}
		if got != ms {
			t.Fatalf("Timestamp = %d, want %d", got, ms)
		}
	}
}

func TestZeroTimestampPart(t *testing.T) {
	entropy := make([]byte, 10)
	s, err := NewWithEntropy(0, entropy)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if s[:10] != "0000000000" {
		t.Fatalf("timestamp part = %q, want 0000000000", s[:10])
	}
	if s != "00000000000000000000000000" {
		t.Fatalf("all-zero ulid = %q", s)
	}
}

func TestLexicalOrder(t *testing.T) {
	entropy := make([]byte, 10)
	small, _ := NewWithEntropy(1000, entropy)
	large, _ := NewWithEntropy(2000, entropy)
	if !(small < large) {
		t.Fatalf("expected %q < %q", small, large)
	}
}

func TestDecodeRoundTrip(t *testing.T) {
	entropy := []byte{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	s, _ := NewWithEntropy(1469918176385, entropy)
	b, err := Decode(s)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if encode(b) != s {
		t.Fatalf("round trip = %q, want %q", encode(b), s)
	}
	for i, e := range entropy {
		if b[6+i] != e {
			t.Fatalf("entropy byte %d = %d, want %d", i, b[6+i], e)
		}
	}
}
