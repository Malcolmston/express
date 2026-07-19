package randomstring

import (
	"regexp"
	"testing"
)

// Upstream parity tests for the npm "randomstring" library (klughammer).
//
// Vectors are transcribed from the original project's real test suite and
// library sources:
//
//	https://raw.githubusercontent.com/klughammer/node-randomstring/master/test/index.js
//	https://raw.githubusercontent.com/klughammer/node-randomstring/master/lib/charset.js
//	https://raw.githubusercontent.com/klughammer/node-randomstring/master/lib/randomstring.js
//
// Because the library emits random data, upstream asserts invariants rather
// than fixed values: exact output length, and that every character falls within
// the allowed set for the requested charset / options. These tests mirror those
// invariants against the Go port. testLength matches upstream's own testLength.
const testLength = 24700

// "defaults to 32 characters in length" and returns alphanumeric.
func TestParityDefaultLength(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 32 {
		t.Fatalf("default length = %d, want 32", len(s))
	}
	if regexp.MustCompile(`[^0-9A-Za-z]`).MatchString(s) {
		t.Fatalf("default output has non-alphanumeric char: %q", s)
	}
}

// "accepts length as an optional first argument": random(10) -> 10, random(0) -> 0.
func TestParityLengthArgument(t *testing.T) {
	for _, n := range []int{10, 0, 7} {
		s, err := Generate(n, "")
		if err != nil {
			t.Fatalf("Generate(%d): %v", n, err)
		}
		if len(s) != n {
			t.Fatalf("Generate(%d) length = %d, want %d", n, len(s), n)
		}
	}
}

// "accepts 'numeric' as charset option": only digits.
func TestParityNumeric(t *testing.T) {
	assertCharset(t, "numeric", `\D`)
}

// "accepts 'alphabetic' as charset option": no digits.
func TestParityAlphabetic(t *testing.T) {
	assertCharset(t, "alphabetic", `\d`)
}

// "accepts 'hex' as charset option": only [0-9a-f].
func TestParityHex(t *testing.T) {
	assertCharset(t, "hex", `[^0-9a-f]`)
}

// "accepts 'binary' as charset option": only [01].
func TestParityBinary(t *testing.T) {
	assertCharset(t, "binary", `[^01]`)
}

// "accepts 'octal' as charset option": only [0-7].
func TestParityOctal(t *testing.T) {
	assertCharset(t, "octal", `[^0-7]`)
}

// assertCharset generates testLength chars from a preset and asserts the length
// and that no character matches the forbidden pattern.
func assertCharset(t *testing.T, preset, forbidden string) {
	t.Helper()
	s, err := Generate(testLength, preset)
	if err != nil {
		t.Fatalf("Generate(%d, %q): %v", testLength, preset, err)
	}
	if len(s) != testLength {
		t.Fatalf("%s length = %d, want %d", preset, len(s), testLength)
	}
	if regexp.MustCompile(forbidden).MatchString(s) {
		t.Fatalf("%s produced forbidden char (pattern %s) in output", preset, forbidden)
	}
}

// "accepts custom charset": charset "abc" -> only [abc].
func TestParityCustomCharset(t *testing.T) {
	s, err := GenerateFrom(testLength, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != testLength {
		t.Fatalf("custom length = %d, want %d", len(s), testLength)
	}
	if regexp.MustCompile(`[^abc]`).MatchString(s) {
		t.Fatalf("custom charset produced out-of-range char")
	}
}

// "accepts an array of charsets": ['alphabetic', '!'] -> only [A-Za-z!].
func TestParityArrayCharset(t *testing.T) {
	s, err := GenerateWith(Options{Length: testLength, Charset: []string{"alphabetic", "!"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != testLength {
		t.Fatalf("array charset length = %d, want %d", len(s), testLength)
	}
	if regexp.MustCompile(`[^A-Za-z!]`).MatchString(s) {
		t.Fatalf("array charset produced out-of-range char")
	}
}

// "accepts readable option": output has no ambiguous chars 0, O, I, l.
func TestParityReadable(t *testing.T) {
	s, err := GenerateWith(Options{Length: testLength, Readable: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != testLength {
		t.Fatalf("readable length = %d, want %d", len(s), testLength)
	}
	if regexp.MustCompile(`[0OIl]`).MatchString(s) {
		t.Fatalf("readable option produced an ambiguous char")
	}
}

// "accepts 'uppercase' as capitalization option": no [a-z].
func TestParityUppercase(t *testing.T) {
	s, err := GenerateWith(Options{Length: testLength, Capitalization: "uppercase"})
	if err != nil {
		t.Fatal(err)
	}
	if regexp.MustCompile(`[a-z]`).MatchString(s) {
		t.Fatalf("uppercase capitalization produced a lowercase char")
	}
}

// "accepts 'lowercase' as capitalization option": no [A-Z].
func TestParityLowercase(t *testing.T) {
	s, err := GenerateWith(Options{Length: testLength, Capitalization: "lowercase"})
	if err != nil {
		t.Fatal(err)
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(s) {
		t.Fatalf("lowercase capitalization produced an uppercase char")
	}
}

// "returns unique strings": 1000 default draws are all distinct.
func TestParityUniqueStrings(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		s, err := New()
		if err != nil {
			t.Fatal(err)
		}
		if seen[s] {
			t.Fatalf("duplicate string produced: %q", s)
		}
		seen[s] = true
	}
}
