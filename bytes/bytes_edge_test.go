package bytes

import (
	"math"
	"strings"
	"testing"
)

// TestFormatUnitFallback pins the upstream rule that an unrecognized `unit`
// option is ignored in favour of auto-selection (npm: `if (!unit ||
// !map[unit.toLowerCase()])`). This port previously divided by 1 and emitted the
// bogus unit verbatim.
func TestFormatUnitFallback(t *testing.T) {
	cases := []struct {
		name string
		n    int64
		unit string
		want string
	}{
		{"unknown unit falls back to auto", 1 << 20, "bananas", "1MB"},
		{"empty unit auto-selects", 1 << 20, "", "1MB"},
		{"unknown unit on small value", 500, "zz", "500B"},
		{"valid unit is honoured", 1 << 20, "KB", "1024KB"},
		{"valid unit keeps caller casing", 1 << 20, "kb", "1024kb"},
		{"valid unit mixed casing", 1 << 30, "Mb", "1024Mb"},
		{"forcing B", 2048, "B", "2048B"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatOpts(tc.n, FormatOptions{Unit: tc.unit}); got != tc.want {
				t.Errorf("FormatOpts(%d, Unit=%q) = %q, want %q", tc.n, tc.unit, got, tc.want)
			}
		})
	}
}

// TestParseOverflow checks that a value too large for an int64 is reported
// rather than silently converted; Go leaves out-of-range float64 -> int64
// conversions implementation-defined.
func TestParseOverflow(t *testing.T) {
	for _, in := range []string{"99999999999999999999pb", "1000000000000000000000", "-99999999999999999999pb"} {
		if got, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) = %d, want an overflow error", in, got)
		} else if !strings.Contains(err.Error(), "overflow") {
			t.Errorf("Parse(%q) error = %v, want an overflow error", in, err)
		}
	}
}

// TestParseBoundary verifies values right at the edge of the int64 range still
// parse, so the overflow guard is not over-eager.
func TestParseBoundary(t *testing.T) {
	if got, err := Parse("9007199254740992"); err != nil || got != 9007199254740992 {
		t.Errorf("Parse(2^53) = %d, %v", got, err)
	}
	if got, err := Parse("8191pb"); err != nil || got != 8191*PB {
		t.Errorf("Parse(8191pb) = %d, %v; want %d", got, err, 8191*PB)
	}
	if _, err := Parse("8192pb"); err == nil {
		t.Error("Parse(8192pb) should overflow int64 (8192 * 2^50 == 2^63)")
	}
}

// TestParseFloorSemantics documents that the multiplication is floored toward
// negative infinity, matching JavaScript's Math.floor.
func TestParseFloorSemantics(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"1.5kb", 1536},
		{"-1.5kb", -1536},
		{"1.9", 1},
		// A bare decimal never reaches the floor: the unit grammar requires a
		// suffix, so parseInt truncates at the dot. Upstream answers -1 here, not
		// -2, which is what this case asserted before the parse path was aligned
		// with bytes@3.1.2.
		{"-1.9", -1},
		{"+1.5kb", 1536},
		{"0", 0},
		{"0kb", 0},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Parse(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// TestParseInvalid covers the error paths. Upstream's only failure condition is
// isNaN(parseInt(val, 10)) once the unit grammar has missed, i.e. "there is no
// leading decimal integer", so nothing else belongs here. "1zb", "1.2.3kb",
// "1 2 kb" and "1e3" used to be asserted as errors and are not: they take the
// parseInt fallback, which is what TestParseIntFallback now pins.
func TestParseInvalid(t *testing.T) {
	for _, in := range []string{"", "   ", "\n", "abc", "kb", "b", "-", "+", ".", ".5", "--1kb", "+-1", "NaN", "Infinity", "kb1", "_1024"} {
		if got, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) = %d, want an error", in, got)
		}
	}
}

// TestParseIntFallback pins the fallback upstream applies to every string its
// unit grammar (/^((-|\+)?(\d+(?:\.\d+)?)) *(kb|mb|gb|tb|pb)$/i) does not match:
// parseInt(value, 10), reading a leading decimal integer as a count of BYTES and
// discarding the rest. It is a wide door — an unknown unit, a misplaced space or
// any trailing text is silently reinterpreted rather than rejected — but it is
// upstream's door, and the parity harness in parity/express/nested/bytes
// measures every vector below against bytes@3.1.2.
func TestParseIntFallback(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"1b", 1},          // "b" is absent from the unit alternation
		{"1B", 1},          //
		{"1.5b", 1},        // parseInt stops at the dot
		{"-1.5b", -1},      // truncation toward zero, not a floor
		{"1zb", 1},         // unknown unit: not an error, just ignored
		{"1kib", 1},        //
		{"1 byte", 1},      //
		{"5 kilobytes", 5}, // the unit the caller wrote is thrown away
		{"1024 please", 1024},
		{"1\tkb", 1},   // a tab is not the space the grammar allows
		{" 1kb", 1},    // the grammar is anchored: leading space demotes to bytes
		{"1kb ", 1},    // and so does a trailing one
		{"\n1kb", 1},   //
		{"4 2", 4},     //
		{"1e3", 1},     // parseInt is radix 10 only
		{"0x10", 0},    // ... so it reads the leading 0 and stops
		{"1,024", 1},   //
		{"100%", 100},  //
		{"1.2.3kb", 1}, //
		{"1e999", 1},   // no Infinity can escape this path
		{"  42", 42},   // leading whitespace is skipped, as parseInt does
		{"42  ", 42},
		{"007", 7},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Parse(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// TestFormatToFixedTies pins that rounding follows Number.prototype.toFixed,
// which breaks a tie away from zero. Go's strconv.FormatFloat rounds a tie to
// even and would print "1.12KB" for 1152 bytes.
func TestFormatToFixedTies(t *testing.T) {
	zero := 0
	cases := []struct {
		n    int64
		opts FormatOptions
		want string
	}{
		{1152, FormatOptions{}, "1.13KB"},
		{-1152, FormatOptions{}, "-1.13KB"},
		{3200, FormatOptions{}, "3.13KB"},
		{1408, FormatOptions{}, "1.38KB"},
		{2560, FormatOptions{DecimalPlaces: &zero}, "3KB"},
		{-2560, FormatOptions{DecimalPlaces: &zero}, "-3KB"},
		{-100, FormatOptions{DecimalPlaces: &zero, Unit: "KB"}, "-0KB"},
	}
	for _, tc := range cases {
		if got := FormatOpts(tc.n, tc.opts); got != tc.want {
			t.Errorf("FormatOpts(%d, %+v) = %q, want %q", tc.n, tc.opts, got, tc.want)
		}
	}
}

// TestFormatTrimQuirk pins upstream's trailing-zero trimmer, which is a regexp
// (/(?:\.0*|(\.[^0]+)0+)$/) rather than a general "strip trailing zeros": a
// fractional part that starts with a zero matches neither branch and keeps its
// zeros.
func TestFormatTrimQuirk(t *testing.T) {
	three, four := 3, 4
	cases := []struct {
		n    int64
		opts FormatOptions
		want string
	}{
		{1034, FormatOptions{DecimalPlaces: &three}, "1.010KB"}, // NOT "1.01KB"
		{1034, FormatOptions{DecimalPlaces: &four}, "1.0098KB"},
		{1536, FormatOptions{DecimalPlaces: &three}, "1.5KB"}, // ".500" does trim
		{2048, FormatOptions{}, "2KB"},                        // ".00" trims wholly
	}
	for _, tc := range cases {
		if got := FormatOpts(tc.n, tc.opts); got != tc.want {
			t.Errorf("FormatOpts(%d, %+v) = %q, want %q", tc.n, tc.opts, got, tc.want)
		}
	}
}

// TestFormatDecimalPlacesClamped documents the one place FormatOpts cannot
// follow upstream: toFixed throws a RangeError outside 0..100 and a Go function
// returning a plain string has nowhere to put that, so the value is clamped.
func TestFormatDecimalPlacesClamped(t *testing.T) {
	neg, big := -1, 101
	if got := FormatOpts(1536, FormatOptions{DecimalPlaces: &neg}); got != "2KB" {
		t.Errorf("FormatOpts(1536, DecimalPlaces=-1) = %q, want %q", got, "2KB")
	}
	if got := FormatOpts(1536, FormatOptions{DecimalPlaces: &big}); got != "1.5KB" {
		t.Errorf("FormatOpts(1536, DecimalPlaces=101) = %q, want %q", got, "1.5KB")
	}
}

// TestFormatRoundTrip checks that formatting then parsing returns the original
// magnitude for exact powers of the unit ladder.
func TestFormatRoundTrip(t *testing.T) {
	for _, n := range []int64{0, 1, 512, KB, 3 * KB, MB, GB, TB, PB, 5 * PB} {
		s := Format(n)
		got, err := Parse(s)
		if err != nil {
			t.Errorf("Parse(Format(%d)=%q) error: %v", n, s, err)
			continue
		}
		if got != n {
			t.Errorf("round trip of %d via %q = %d", n, s, got)
		}
	}
}

// TestFormatNegativeAndZero pins the documented edge cases.
func TestFormatNegativeAndZero(t *testing.T) {
	cases := []struct {
		n    int64
		opts FormatOptions
		want string
	}{
		{0, FormatOptions{}, "0B"},
		{-1024, FormatOptions{}, "-1KB"},
		{-1536, FormatOptions{}, "-1.5KB"},
		{math.MinInt64, FormatOptions{}, "-8192PB"},
		{1024, FormatOptions{FixedDecimals: true}, "1.00KB"},
		{1000000, FormatOptions{ThousandsSeparator: ","}, "976.56KB"},
		{-1234567, FormatOptions{Unit: "B", ThousandsSeparator: ","}, "-1,234,567B"},
	}
	for _, tc := range cases {
		if got := FormatOpts(tc.n, tc.opts); got != tc.want {
			t.Errorf("FormatOpts(%d, %+v) = %q, want %q", tc.n, tc.opts, got, tc.want)
		}
	}
}
