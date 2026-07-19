package truncate

import "testing"

// Parity vectors transcribed verbatim from the lodash upstream test suite:
//
//	Source: https://raw.githubusercontent.com/lodash/lodash/master/test/test.js
//	Module: QUnit.module('lodash.truncate') (around line 22757)
//	Fixture string: 'hi-diddly-ho there, neighborino'  (31 code points)
//
// lodash's _.truncate(string, options) accepts an options object with
//   length   (default 30),
//   omission (default '...'),
//   separator (string or RegExp).
// This Go port exposes the same behavior through Truncate(s, length) and
// TruncateOpts(s, length, Options{Ellipsis, WordBoundary}); the vectors below
// map each lodash assertion onto those calls. The lodash `separator: ' '`
// assertion maps onto WordBoundary=true (whitespace boundary). The RegExp
// separator assertions have no port equivalent and are documented as a gap.

const parityString = "hi-diddly-ho there, neighborino"

func TestParityDefaultLength30(t *testing.T) {
	// _.truncate(string) => 'hi-diddly-ho there, neighbo...' (default length 30, omission '...')
	got := TruncateOpts(parityString, 30, Options{Ellipsis: "..."})
	want := "hi-diddly-ho there, neighbo..."
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParityNotTruncatedWhenWithinLength(t *testing.T) {
	// _.truncate(string, {length: string.length})     => string
	// _.truncate(string, {length: string.length + 2}) => string
	if got := TruncateOpts(parityString, len([]rune(parityString)), Options{Ellipsis: "..."}); got != parityString {
		t.Fatalf("length==len: got %q, want %q", got, parityString)
	}
	if got := TruncateOpts(parityString, len([]rune(parityString))+2, Options{Ellipsis: "..."}); got != parityString {
		t.Fatalf("length==len+2: got %q, want %q", got, parityString)
	}
}

func TestParityGivenLength24(t *testing.T) {
	// _.truncate(string, {length: 24}) => 'hi-diddly-ho there, n...'
	got := TruncateOpts(parityString, 24, Options{Ellipsis: "..."})
	want := "hi-diddly-ho there, n..."
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParityOmissionOption(t *testing.T) {
	// _.truncate(string, {omission: ' [...]'}) => 'hi-diddly-ho there, neig [...]' (default length 30)
	got := TruncateOpts(parityString, 30, Options{Ellipsis: " [...]"})
	want := "hi-diddly-ho there, neig [...]"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParityLengthOption4(t *testing.T) {
	// _.truncate(string, {length: 4}) => 'h...'
	got := TruncateOpts(parityString, 4, Options{Ellipsis: "..."})
	want := "h..."
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParityStringSeparator(t *testing.T) {
	// _.truncate(string, {length: 24, separator: ' '}) => 'hi-diddly-ho there,...'
	// Mapped onto WordBoundary=true, whose whitespace boundary matches the ' ' separator.
	got := TruncateOpts(parityString, 24, Options{Ellipsis: "...", WordBoundary: true})
	want := "hi-diddly-ho there,..."
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParityNegativeLengthAsZero(t *testing.T) {
	// _.truncate(string, {length: 0})  => '...'
	// _.truncate(string, {length: -2}) => '...'
	for _, length := range []int{0, -2} {
		got := TruncateOpts(parityString, length, Options{Ellipsis: "..."})
		want := "..."
		if got != want {
			t.Fatalf("length %d: got %q, want %q", length, got, want)
		}
	}
}
