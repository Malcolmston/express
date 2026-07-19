package bytes

import "testing"

// Upstream-parity tests for the npm "bytes" package (visionmedia/bytes.js,
// v3.1.2). Every vector below is taken from the ORIGINAL library, not invented:
//
//   - test/bytes.js  (the actual mocha suite):
//     https://raw.githubusercontent.com/visionmedia/bytes.js/master/test/bytes.js
//   - Readme.md      (documented input -> output examples):
//     https://raw.githubusercontent.com/visionmedia/bytes.js/master/Readme.md
//   - index.js       (unit magnitudes and parse/format behavior):
//     https://raw.githubusercontent.com/visionmedia/bytes.js/master/index.js
//
// The upstream default export bytes(value) dispatches: a string parses to a
// number (bytes.parse), a number formats to a string (bytes.format), and any
// invalid input returns null. In this Go port those map to Parse (string ->
// (int64, error)) and Format/FormatOpts (int64 -> string); a null result maps
// to a Parse error.

// TestParityParse covers vectors asserted or documented upstream for
// bytes.parse / bytes('<string>').
//
// Sources:
//
//	test/bytes.js:  assert.equal(bytes('1KB'), 1024)
//	Readme.md:      bytes.parse('1KB') -> 1024, bytes.parse('1024') -> 1024,
//	                "1TB" (1099511627776), units are powers of two and
//	                case-insensitive.
//	index.js:       map = {b:1, kb:1<<10, mb:1<<20, gb:1<<30, tb:1024^4, pb:1024^5}
func TestParityParse(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"1KB", 1024},          // test/bytes.js constructor assertion
		{"1kb", 1024},          // case-insensitive (Readme: units case-insensitive)
		{"1024", 1024},         // Readme: bytes.parse('1024') -> 1024 (bare number = bytes)
		{"1TB", 1099511627776}, // Readme header: "1TB" to 1099511627776
		{"1MB", 1 << 20},       // index.js map
		{"1GB", 1 << 30},       // index.js map
		{"1PB", 1 << 50},       // index.js map (pb = 1024^5)
		{"1.5MB", 1572864},     // index.js: Math.floor(map.mb * 1.5)
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("Parse(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// TestParityParseInvalid covers upstream's "return null upon error" for
// invalid string input. In this port a null result is signalled by a non-nil
// error.
//
// Source: test/bytes.js -> assert.strictEqual(bytes('foobar'), null)
func TestParityParseInvalid(t *testing.T) {
	if _, err := Parse("foobar"); err == nil {
		t.Errorf("Parse(%q) = nil error, want error (upstream returns null)", "foobar")
	}
}

// TestParityFormat covers vectors asserted or documented upstream for
// bytes.format / bytes(<number>).
//
// Sources:
//
//	test/bytes.js:  assert.equal(bytes(1024), '1KB')
//	Readme.md:      bytes.format(1024) -> '1KB', bytes.format(1000) -> '1000B'
//	index.js:       auto unit selection by magnitude.
func TestParityFormat(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{1024, "1KB"},   // test/bytes.js + Readme
		{1000, "1000B"}, // Readme: bytes.format(1000) -> '1000B' (1000 < 1024)
		{1 << 30, "1GB"},
		{1 << 40, "1TB"},
		{1 << 50, "1PB"},
	}
	for _, c := range cases {
		if got := Format(c.in); got != c.want {
			t.Errorf("Format(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestParityFormatOptions covers the documented bytes.format option examples.
//
// Source: Readme.md
//
//	bytes.format(1000, {thousandsSeparator: ' '}) -> '1 000B'
//	bytes.format(1024 * 1.7, {decimalPlaces: 0})  -> '2KB'
//	bytes.format(1024, {unitSeparator: ' '})      -> '1 KB'
//
// and test/bytes.js:
//
//	bytes(1000, {thousandsSeparator: ' '}) -> '1 000B'
//
// The decimalPlaces example uses 1024*1.7 = 1740.8 in JS; this port formats
// int64 counts, so the equivalent input 1740 rounds identically to "2KB"
// (1740/1024 = 1.699..., toFixed(0) = "2").
func TestParityFormatOptions(t *testing.T) {
	zero := 0

	if got := FormatOpts(1000, FormatOptions{ThousandsSeparator: " "}); got != "1 000B" {
		t.Errorf("FormatOpts(1000, thousandsSeparator=' ') = %q, want %q", got, "1 000B")
	}
	if got := FormatOpts(1740, FormatOptions{DecimalPlaces: &zero}); got != "2KB" {
		t.Errorf("FormatOpts(1740, decimalPlaces=0) = %q, want %q", got, "2KB")
	}
	if got := FormatOpts(1024, FormatOptions{UnitSeparator: " "}); got != "1 KB" {
		t.Errorf("FormatOpts(1024, unitSeparator=' ') = %q, want %q", got, "1 KB")
	}
}

// TestParityRoundTrip mirrors the upstream Readme headline claim that parsing
// and formatting are inverses for the documented pairs.
//
// Source: Readme.md
//
//	bytes(1024) -> '1KB' and bytes('1KB') -> 1024
func TestParityRoundTrip(t *testing.T) {
	if got := Format(1024); got != "1KB" {
		t.Fatalf("Format(1024) = %q, want %q", got, "1KB")
	}
	n, err := Parse("1KB")
	if err != nil {
		t.Fatalf("Parse(%q): %v", "1KB", err)
	}
	if n != 1024 {
		t.Fatalf("Parse(%q) = %d, want 1024", "1KB", n)
	}
}
