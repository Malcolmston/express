package uidsafe

import (
	"encoding/base64"
	"strings"
	"testing"
)

// Parity tests for the npm "uid-safe" library (crypto-utils/uid-safe).
//
// Vectors transcribed from the canonical upstream test suite and source:
//   https://raw.githubusercontent.com/crypto-utils/uid-safe/master/test/test.js
//   https://raw.githubusercontent.com/crypto-utils/uid-safe/master/index.js
//
// uid-safe reads `length` random bytes and encodes them as base64 with the
// trailing '=' padding stripped and '+'/'/' replaced by '-'/'_' — i.e. exactly
// Go's base64.RawURLEncoding. The library's outputs are random, so the upstream
// tests assert structural invariants rather than fixed strings:
//
//   - uid(18)/uid.sync(18)  => Buffer.byteLength(val) === 24   (length invariant)
//   - large inputs          => no '+', '/', or '=' in output   (url-safe, unpadded)
//   - charset               => [A-Za-z0-9_-]
//
// This port exposes Bytes(n) (error-returning) and MustBytes(n) (panicking) in
// place of uid.sync/uid; both take a byte count and return the url-safe string.

const urlSafeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

// rawURLEncodedLen is the exact length base64.RawURLEncoding produces for n
// input bytes: ceil(n*8/6). It is the invariant upstream asserts via
// Buffer.byteLength.
func rawURLEncodedLen(n int) int { return (n*8 + 5) / 6 }

// TestParityLengthInvariant checks the upstream "correct length" assertion for a
// range of byte counts, including the canonical uid(18) === 24 case.
func TestParityLengthInvariant(t *testing.T) {
	cases := []struct {
		n    int
		want int
	}{
		{0, 0},
		{1, 2},
		{3, 4},
		{15, 20}, // MustBytes example
		{18, 24}, // canonical uid.sync(18) === 24
		{24, 32},
		{100, 134},
	}
	for _, c := range cases {
		if got := rawURLEncodedLen(c.n); got != c.want {
			t.Fatalf("encoded-len formula wrong for n=%d: got %d want %d", c.n, got, c.want)
		}
		s, err := Bytes(c.n)
		if err != nil {
			t.Fatalf("Bytes(%d) error: %v", c.n, err)
		}
		if len(s) != c.want {
			t.Fatalf("Bytes(%d) length = %d, want %d", c.n, len(s), c.want)
		}
		if len(MustBytes(c.n)) != c.want {
			t.Fatalf("MustBytes(%d) length = %d, want %d", c.n, len(MustBytes(c.n)), c.want)
		}
	}
}

// TestParityNoUnsafeChars mirrors the upstream "should not contain +, /, or ="
// assertion over a large input, ensuring the output is url-safe and unpadded.
func TestParityNoUnsafeChars(t *testing.T) {
	s, err := Bytes(100000)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(s, "+/=") {
		t.Fatalf("output contains +, /, or =: %.40q...", s)
	}
	if want := rawURLEncodedLen(100000); len(s) != want {
		t.Fatalf("Bytes(100000) length = %d, want %d", len(s), want)
	}
}

// TestParityCharset asserts every output character is in the url-safe base64
// alphabet [A-Za-z0-9_-].
func TestParityCharset(t *testing.T) {
	s, err := Bytes(4096)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(s); i++ {
		if !strings.ContainsRune(urlSafeAlphabet, rune(s[i])) {
			t.Fatalf("byte %d = %q not in url-safe alphabet", i, s[i])
		}
	}
}

// TestParityRoundTrip confirms the output decodes back to exactly n bytes under
// RawURLEncoding — the encoding uid-safe's toString() is equivalent to.
func TestParityRoundTrip(t *testing.T) {
	for _, n := range []int{0, 1, 18, 24, 33} {
		s, err := Bytes(n)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := base64.RawURLEncoding.DecodeString(s)
		if err != nil {
			t.Fatalf("Bytes(%d) not RawURLEncoding-decodable: %v", n, err)
		}
		if len(raw) != n {
			t.Fatalf("Bytes(%d) decoded to %d bytes", n, len(raw))
		}
	}
}
