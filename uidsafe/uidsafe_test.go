package uidsafe

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestBytesURLSafe(t *testing.T) {
	s, err := Bytes(18)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(s, "+/=") {
		t.Fatalf("result not url-safe or padded: %q", s)
	}
}

func TestBytesDecodable(t *testing.T) {
	s, err := Bytes(18)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(raw) != 18 {
		t.Fatalf("expected 18 bytes, got %d", len(raw))
	}
}

func TestBytesUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		s, err := Bytes(18)
		if err != nil {
			t.Fatal(err)
		}
		if seen[s] {
			t.Fatalf("collision detected: %q", s)
		}
		seen[s] = true
	}
}

func TestBytesLength(t *testing.T) {
	// RawURLEncoding length is ceil(n*8/6).
	s, err := Bytes(24)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 32 {
		t.Fatalf("expected length 32 for 24 bytes, got %d", len(s))
	}
}

func TestBytesZero(t *testing.T) {
	s, err := Bytes(0)
	if err != nil {
		t.Fatal(err)
	}
	if s != "" {
		t.Fatalf("expected empty string, got %q", s)
	}
}

func TestMustBytes(t *testing.T) {
	s := MustBytes(18)
	if s == "" {
		t.Fatal("expected non-empty string")
	}
}
