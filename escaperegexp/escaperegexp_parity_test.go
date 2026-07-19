package escaperegexp

import (
	"regexp"
	"testing"
)

// Parity tests transcribed verbatim from the upstream npm package
// sindresorhus/escape-string-regexp. Input -> expected-output vectors are the
// real values from the upstream test suite, not invented.
//
// Upstream sources (fetched 2026-07-19):
//   https://raw.githubusercontent.com/sindresorhus/escape-string-regexp/main/test.js
//   https://raw.githubusercontent.com/sindresorhus/escape-string-regexp/main/index.js
//
// Upstream index.js logic:
//   string
//     .replace(/[|\\{}()[\]^$+*?.]/g, '\\$&')
//     .replace(/-/g, '\\x2d');

// TestParityMain mirrors upstream test('main', ...):
//
//	escapeStringRegexp('\\ ^ $ * + ? . ( ) | { } [ ]')
//	  === '\\\\ \\^ \\$ \\* \\+ \\? \\. \\( \\) \\| \\{ \\} \\[ \\]'
//
// The JS string literals unescape to the Go raw strings below.
func TestParityMain(t *testing.T) {
	in := `\ ^ $ * + ? . ( ) | { } [ ]`
	want := `\\ \^ \$ \* \+ \? \. \( \) \| \{ \} \[ \]`
	if got := EscapeRegExp(in); got != want {
		t.Fatalf("EscapeRegExp(%q) = %q, want %q", in, got, want)
	}
}

// TestParityHyphenPCRE mirrors upstream
// test('escapes `-` in a way compatible with PCRE', ...):
//
//	escapeStringRegexp('foo - bar') === 'foo \\x2d bar'
func TestParityHyphenPCRE(t *testing.T) {
	in := "foo - bar"
	want := `foo \x2d bar`
	if got := EscapeRegExp(in); got != want {
		t.Fatalf("EscapeRegExp(%q) = %q, want %q", in, got, want)
	}
}

// TestParityHyphenUnicodeFlag mirrors upstream
// test('escapes `-` in a way compatible with the Unicode flag', ...):
//
//	new RegExp(escapeStringRegexp('-'), 'u') matches '-'
//
// Go's regexp (RE2) understands \x2d, so the escaped pattern must compile and
// match a literal hyphen.
func TestParityHyphenUnicodeFlag(t *testing.T) {
	escaped := EscapeRegExp("-")
	if escaped != `\x2d` {
		t.Fatalf("EscapeRegExp(%q) = %q, want %q", "-", escaped, `\x2d`)
	}
	re, err := regexp.Compile(escaped)
	if err != nil {
		t.Fatalf("regexp.Compile(%q) error: %v", escaped, err)
	}
	if !re.MatchString("-") {
		t.Fatalf("escaped pattern %q did not match %q", escaped, "-")
	}
}
