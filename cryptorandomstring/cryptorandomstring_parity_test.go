package cryptorandomstring

// Upstream parity tests for the npm library "sindresorhus/crypto-random-string".
//
// Canonical character sets and type->charset mapping taken verbatim from the
// upstream source and test suite:
//   - core.js: https://raw.githubusercontent.com/sindresorhus/crypto-random-string/main/core.js
//   - test.js: https://raw.githubusercontent.com/sindresorhus/crypto-random-string/main/test.js
//
// From core.js the exact charset constants are:
//   numeric:         "0123456789"
//   distinguishable: "CDEHKMPRTUWXY012458"
//   url-safe:        "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
//   ascii-printable: "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
//   alphanumeric:    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
// hex ("0123456789abcdef") and base64 (standard A-Za-z0-9+/) come from Node's
// randomBytes(...).toString('hex'|'base64'), matched by test.js patterns
// /^[a-f\d]*$/ and /^[a-zA-Z\d/+]*$/.
//
// Because outputs are random, these vectors assert (a) the exact canonical
// charset for each type/preset, (b) the type->charset mapping, and (c) the
// length + charset-membership invariants the upstream test suite checks.

import (
	"sort"
	"strings"
	"testing"
)

// canonical charsets straight from upstream core.js / test.js.
const (
	canonHex             = "0123456789abcdef"
	canonBase64          = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	canonURLSafe         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	canonNumeric         = "0123456789"
	canonDistinguishable = "CDEHKMPRTUWXY012458"
	canonAsciiPrintable  = "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	canonAlphanumeric    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

func sortedRunes(s string) string {
	r := []rune(s)
	sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })
	return string(r)
}

// setEqual reports whether two strings contain the exact same set of runes,
// ignoring order (order is irrelevant for a uniformly-sampled charset).
func setEqual(a, b string) bool {
	return sortedRunes(a) == sortedRunes(b)
}

// TestParityCharsetConstants asserts every preset's internal charset constant
// exactly matches the canonical upstream character set.
func TestParityCharsetConstants(t *testing.T) {
	cases := []struct {
		name  string
		got   string
		want  string
		exact bool // require identical order too
	}{
		{"hex", hexChars, canonHex, true},
		{"base64", base64Chars, canonBase64, true},
		{"url-safe", urlSafeChars, canonURLSafe, false},
		{"numeric", numericChars, canonNumeric, true},
		{"distinguishable", distinguishableChars, canonDistinguishable, false},
		{"ascii-printable", asciiPrintableChars, canonAsciiPrintable, true},
		{"alphanumeric", alphanumericChars, canonAlphanumeric, true},
	}
	for _, c := range cases {
		if c.exact {
			if c.got != c.want {
				t.Errorf("%s charset = %q, want %q", c.name, c.got, c.want)
			}
			continue
		}
		if !setEqual(c.got, c.want) {
			t.Errorf("%s charset (as a set) = %q, want %q", c.name, sortedRunes(c.got), sortedRunes(c.want))
		}
	}
}

// TestParityTypeMapping asserts charsForType maps each upstream type name to
// the canonical charset (default type "" and "hex" both -> hex).
func TestParityTypeMapping(t *testing.T) {
	cases := []struct {
		typ  string
		want string
	}{
		{"", canonHex},
		{"hex", canonHex},
		{"base64", canonBase64},
		{"url-safe", canonURLSafe},
		{"numeric", canonNumeric},
		{"distinguishable", canonDistinguishable},
		{"ascii-printable", canonAsciiPrintable},
		{"alphanumeric", canonAlphanumeric},
	}
	for _, c := range cases {
		got, err := charsForType(c.typ)
		if err != nil {
			t.Errorf("charsForType(%q): unexpected error %v", c.typ, err)
			continue
		}
		if !setEqual(got, c.want) {
			t.Errorf("charsForType(%q) = %q, want set %q", c.typ, got, c.want)
		}
	}
}

// TestParityDistinguishableExcludesAmbiguous verifies the distinguishable set
// omits the visually ambiguous glyphs the upstream set excludes (e.g. 3,6,7,9,
// A,B,F,...). This directly guards against the previously-present spurious
// "964" suffix in the port's constant.
func TestParityDistinguishableExcludesAmbiguous(t *testing.T) {
	forbidden := "9673GS" // none of these appear in "CDEHKMPRTUWXY012458"
	for _, r := range forbidden {
		if strings.ContainsRune(distinguishableChars, r) {
			t.Errorf("distinguishable charset must not contain %q", r)
		}
	}
	// Sampled output must never contain a forbidden glyph.
	s, err := Generate(Options{Length: 2000, Type: "distinguishable"})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range s {
		if !strings.ContainsRune(canonDistinguishable, r) {
			t.Fatalf("distinguishable produced out-of-set char %q", r)
		}
	}
}

// TestParityURLSafeIncludesDotTilde verifies url-safe charset includes '.' and
// '~' per upstream (pattern /^[\w.~-]*$/, 66 chars), which the port previously
// omitted.
func TestParityURLSafeIncludesDotTilde(t *testing.T) {
	if !strings.ContainsRune(urlSafeChars, '.') {
		t.Error("url-safe charset must contain '.'")
	}
	if !strings.ContainsRune(urlSafeChars, '~') {
		t.Error("url-safe charset must contain '~'")
	}
	if len([]rune(urlSafeChars)) != 66 {
		t.Errorf("url-safe charset size = %d, want 66", len([]rune(urlSafeChars)))
	}
}

// TestParityCharsetSizes asserts the upstream character-set sizes reported by
// test.js.
func TestParityCharsetSizes(t *testing.T) {
	cases := []struct {
		typ  string
		size int
	}{
		{"hex", 16},
		{"base64", 64},
		{"url-safe", 66},
		{"numeric", 10},
		{"distinguishable", 19},
		{"alphanumeric", 62},
	}
	for _, c := range cases {
		got, err := charsForType(c.typ)
		if err != nil {
			t.Fatalf("charsForType(%q): %v", c.typ, err)
		}
		if n := len([]rune(got)); n != c.size {
			t.Errorf("%s charset size = %d, want %d", c.typ, n, c.size)
		}
	}
}

// TestParityLengthAndMembership mirrors the upstream invariants: output has the
// requested length and every character is drawn from the requested charset.
func TestParityLengthAndMembership(t *testing.T) {
	cases := []struct {
		typ     string
		charset string
	}{
		{"hex", canonHex},
		{"base64", canonBase64},
		{"url-safe", canonURLSafe},
		{"numeric", canonNumeric},
		{"distinguishable", canonDistinguishable},
		{"ascii-printable", canonAsciiPrintable},
		{"alphanumeric", canonAlphanumeric},
	}
	for _, c := range cases {
		for _, length := range []int{0, 1, 10, 100} {
			s, err := Generate(Options{Length: length, Type: c.typ})
			if err != nil {
				t.Fatalf("Generate(%q, %d): %v", c.typ, length, err)
			}
			if len([]rune(s)) != length {
				t.Errorf("Generate(%q, %d) length = %d, want %d", c.typ, length, len([]rune(s)), length)
			}
			for _, r := range s {
				if !strings.ContainsRune(c.charset, r) {
					t.Fatalf("Generate(%q) out-of-set char %q", c.typ, r)
				}
			}
		}
	}
}

// TestParityZeroLength mirrors the upstream case: length 0 yields "".
func TestParityZeroLength(t *testing.T) {
	s, err := Generate(Options{Length: 0, Type: "hex"})
	if err != nil {
		t.Fatal(err)
	}
	if s != "" {
		t.Errorf("length 0 = %q, want empty", s)
	}
}
