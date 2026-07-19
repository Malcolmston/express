package color

// Upstream-parity vectors for the npm library Qix-/color-string ("Parser and
// generator for CSS color strings", v2.1.4). The concrete input -> expected
// values below are transcribed verbatim from the upstream test suite:
//
//	https://raw.githubusercontent.com/Qix-/color-string/master/test.js
//	https://raw.githubusercontent.com/Qix-/color-string/master/index.js
//
// Scope note: this Go package is a color-primitives port whose surface overlaps
// color-string only on hex parsing (upstream string.get.rgb of a hex literal)
// and hex generation (upstream string.to.hex). color-string additionally parses
// rgb()/hsl()/hwb() functional notation, percentage/alpha syntax and CSS named
// colours, and returns a 4-element [r,g,b,a] array; this port exposes none of
// that string-parsing surface and its RGB type carries no alpha channel. Those
// unmapped behaviours are recorded as gaps in the task notes rather than tested
// here. For the mapped hex surface the port matches upstream on the r,g,b
// channels; the only intentional divergence is hex-output casing (upstream
// emits uppercase "#RRGGBB", this port emits lowercase), which the format test
// below compares case-insensitively.

import (
	"strings"
	"testing"
)

// TestParityHexToRGBValid covers upstream `string.get.rgb('<hex>')` for the
// hex-literal forms the port supports: 3-nibble, 4-nibble (rgba), 6-digit and
// 8-digit (rrggbbaa). Upstream returns [r,g,b,a]; the port drops alpha, so only
// the r,g,b channels are asserted.
func TestParityHexToRGBValid(t *testing.T) {
	cases := []struct {
		in      string
		r, g, b uint8
	}{
		// 3-nibble.  upstream: get.rgb('#fef') -> [255,238,255,1]
		{"#fef", 255, 238, 255},
		// 6-digit, mixed case.  get.rgb('#fffFEF') -> [255,255,239,1]
		{"#fffFEF", 255, 255, 239},
		// 8-digit (rrggbbaa).  get.rgb('#fffFEFff') -> [255,255,239,1]
		{"#fffFEFff", 255, 255, 239},
		// 8-digit, zero alpha.  get.rgb('#fffFEF00') -> [255,255,239,0]
		{"#fffFEF00", 255, 255, 239},
		// 8-digit, fractional alpha.  get.rgb('#fffFEFa9') -> [255,255,239,0.66]
		{"#fffFEFa9", 255, 255, 239},
		// 4-nibble (rgba).  get.rgb('#fffa') -> [255,255,255,0.67]
		{"#fffa", 255, 255, 255},
		// 8-digit.  get.rgb('#c814e933') -> [200,20,233,0.2]
		{"#c814e933", 200, 20, 233},
		// 8-digit, zero alpha.  get.rgb('#c814e900') -> [200,20,233,0]
		{"#c814e900", 200, 20, 233},
		// 8-digit, full alpha.  get.rgb('#c814e9ff') -> [200,20,233,1]
		{"#c814e9ff", 200, 20, 233},
	}
	for _, tt := range cases {
		got, err := HexToRGB(tt.in)
		if err != nil {
			t.Errorf("HexToRGB(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if got.R != tt.r || got.G != tt.g || got.B != tt.b {
			t.Errorf("HexToRGB(%q) = %d,%d,%d; upstream rgb = %d,%d,%d",
				tt.in, got.R, got.G, got.B, tt.r, tt.g, tt.b)
		}
	}
}

// TestParityHexToRGBInvalid covers upstream `string.get.rgb('<hex>')` returning
// null for malformed hex literals; the port must return an error for the same
// inputs.
func TestParityHexToRGBInvalid(t *testing.T) {
	// Transcribed from upstream "Invalid" assertions that use a leading '#'.
	// Bare non-'#' strings such as '333333' are excluded: upstream rejects them
	// (its keyword branch requires a named colour), but this port intentionally
	// accepts leading-'#'-optional hex, so that is a documented divergence, not
	// a parity failure.
	invalid := []string{
		"#1",       // 1 nibble
		"#f",       // 1 nibble
		"#4f",      // 2 nibbles
		"#45ab4",   // 5 nibbles
		"#45ab45e", // 7 nibbles
	}
	for _, in := range invalid {
		if _, err := HexToRGB(in); err == nil {
			t.Errorf("HexToRGB(%q) = nil error; upstream get.rgb returns null", in)
		}
	}
}

// TestParityToHex covers upstream `string.to.hex(r, g, b[, a])`. Upstream emits
// uppercase "#RRGGBB" and drops alpha; this port emits lowercase, so the values
// are compared case-insensitively.
func TestParityToHex(t *testing.T) {
	cases := []struct {
		c    RGB
		want string // upstream to.hex output
	}{
		// to.hex(255, 10, 35)    -> '#FF0A23'
		{RGB{255, 10, 35}, "#FF0A23"},
		// to.hex(255, 10, 35, 1) -> '#FF0A23'  (alpha dropped)
		{RGB{255, 10, 35}, "#FF0A23"},
		// to.hex(44.2, 83.8, 44) -> '#2C542C'  (upstream rounds; the port takes
		// pre-rounded uint8 channels, so the rounded values 44,84,44 are used).
		{RGB{44, 84, 44}, "#2C542C"},
	}
	for _, tt := range cases {
		got := RGBToHex(tt.c)
		if !strings.EqualFold(got, tt.want) {
			t.Errorf("RGBToHex(%v) = %q; upstream to.hex = %q (case-insensitive)",
				tt.c, got, tt.want)
		}
	}
}
