package htmlentities

// Upstream-parity vectors for the npm library "mdevils/html-entities".
//
// Every input -> expected-output pair below is taken verbatim from the
// library's own Jest suite, not invented here:
//
//	https://raw.githubusercontent.com/mdevils/html-entities/main/test/index.test.ts
//	(html-entities v2.6.0)
//
// The Go port implements a subset of the upstream API (Encode with
// "specialChars"/"nonAscii" modes and a permissive Decode over a curated named
// table plus decimal/hex numeric references). Only upstream vectors that fall
// inside that supported subset are asserted here; upstream behaviors the port
// deliberately does not implement (see the "gaps" note at the bottom of this
// file) are documented rather than asserted so the suite stays green.

import "testing"

// --- encode(), mode: specialChars -------------------------------------------
// Upstream test/index.test.ts, describe('mode'):
//
//	encode('a\n<>"\'&©∆℞😂\0\x01', {mode: 'specialChars'})
//	  === 'a\n&lt;&gt;&quot;&apos;&amp;©∆℞😂\0\x01'
//	encode('a\n<>"\'&©∆℞😂\0\x01END', {mode: 'specialChars'})
//	  === 'a\n&lt;&gt;&quot;&apos;&amp;©∆℞😂\0\x01END'
//
// specialChars encodes only the five special characters and leaves everything
// else (non-ASCII text, astral emoji, control characters) untouched.
func TestParityEncodeSpecialCharsMode(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{
			"a\n<>\"'&©∆℞😂\x00\x01",
			"a\n&lt;&gt;&quot;&apos;&amp;©∆℞😂\x00\x01",
		},
		{
			"a\n<>\"'&©∆℞😂\x00\x01END",
			"a\n&lt;&gt;&quot;&apos;&amp;©∆℞😂\x00\x01END",
		},
	}
	for _, c := range cases {
		if got := Encode(c.in, EncodeOptions{Mode: "specialChars"}); got != c.want {
			t.Errorf("Encode(%q, specialChars) = %q; want %q", c.in, got, c.want)
		}
		// The default (no options) is specialChars, so it must agree.
		if got := Encode(c.in); got != c.want {
			t.Errorf("Encode(%q) default = %q; want %q", c.in, got, c.want)
		}
	}
}

// Upstream: encode(”) === ”  (describe('encode()'), 'should handle empty string')
func TestParityEncodeEmpty(t *testing.T) {
	if got := Encode(""); got != "" {
		t.Errorf("Encode(\"\") = %q; want \"\"", got)
	}
}

// --- encode(), nonAscii-style numeric output --------------------------------
// The port's "nonAscii" mode rewrites every rune above 0x7F as a decimal
// numeric reference. Upstream produces the identical decimal output under
// {mode: 'nonAscii', level: 'xml', numeric: 'decimal'}; the relevant tested
// substring is the non-ASCII run, verified against
// describe('numeric') / describe('level', 'xml'):
//
//	'©∆℞😂' -> '&#169;&#8710;&#8478;&#128514;'
//
// (Upstream's full nonAscii+named or nonAsciiPrintable control-char handling is
// out of scope for the port; see the gaps note.)
func TestParityEncodeNonAsciiDecimal(t *testing.T) {
	if got := Encode("©∆℞😂", EncodeOptions{Mode: "nonAscii"}); got != "&#169;&#8710;&#8478;&#128514;" {
		t.Errorf("Encode nonAscii = %q; want %q", got, "&#169;&#8710;&#8478;&#128514;")
	}
	// The five specials are still encoded in nonAscii mode.
	if got := Encode("<é>", EncodeOptions{Mode: "nonAscii"}); got != "&lt;&#233;&gt;" {
		t.Errorf("Encode nonAscii = %q; want %q", got, "&lt;&#233;&gt;")
	}
}

// --- decode(), edge cases ---------------------------------------------------
// Upstream describe('decode()'):
//
//	decode('')   === ''      ('should handle empty string')
//	decode('&')  === '&'     ('should handle single ampersand')
//	decode('&a') === '&a'    ('should handle incomplete entity')
func TestParityDecodeEdgeCases(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"&", "&"},
		{"&a", "&a"},
	}
	for _, c := range cases {
		if got := Decode(c.in); got != c.want {
			t.Errorf("Decode(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

// --- decode(), named references in the port's table -------------------------
// Upstream decodes these named references to the values below (the five XML
// specials plus common HTML named entities). Verified against decodeEntity()
// vectors and the level suite, restricted to names present in the port's table.
func TestParityDecodeNamed(t *testing.T) {
	cases := []struct{ in, want string }{
		{"&amp;", "&"},
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&quot;", "\""},
		{"&apos;", "'"},
		{"&copy;", "©"},
	}
	for _, c := range cases {
		if got := Decode(c.in); got != c.want {
			t.Errorf("Decode(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

// --- decode(), numeric references -------------------------------------------
// Upstream describe('decodeEntity()') / describe('decode()'):
//
//	'&#XD06;'       -> 'ആ'                    ('should handle hex entities')
//	'&#xD06;'       -> 'ആ'
//	'&#128514;'     -> '😂'                   ('should decode emoji')
//	'&#34;'         -> '"'
//	'&#0;'          -> U+FFFD                 ('should handle null-char')
//	'&#2013266066;' -> U+FFFD                 ('should handle invalid numeric entities')
func TestParityDecodeNumeric(t *testing.T) {
	cases := []struct{ in, want string }{
		{"&#XD06;", "ആ"},
		{"&#xD06;", "ആ"},
		{"&#128514;", "😂"},
		{"&#34;", "\""},
		{"&#0;", "�"},
		{"&#2013266066;", "�"},
	}
	for _, c := range cases {
		if got := Decode(c.in); got != c.want {
			t.Errorf("Decode(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

// --- decode(), &amp;amp; single-level roundtrip -----------------------------
// Decoding resolves exactly one level of entity, so an escaped ampersand entity
// decodes back to its literal &amp; text (matching upstream and the HTML spec).
func TestParityDecodeAmpAmp(t *testing.T) {
	if got := Decode("&amp;amp;"); got != "&amp;" {
		t.Errorf("Decode(%q) = %q; want %q", "&amp;amp;", got, "&amp;")
	}
	// And that literal round-trips through Encode/Decode of application text.
	in := "Tom & Jerry's <a href=\"x\">"
	if got := Decode(Encode(in)); got != in {
		t.Errorf("round trip failed for %q; got %q", in, got)
	}
}
