package filesize

// Upstream-parity vectors for avoidwork/filesize (the npm package "filesize",
// published from github.com/avoidwork/filesize.js).
//
// Every input -> expected-output pair below is taken verbatim from the upstream
// unit-test suite and source, pinned at v11.0.22:
//
//	https://raw.githubusercontent.com/avoidwork/filesize.js/master/tests/unit/filesize.test.js
//	https://raw.githubusercontent.com/avoidwork/filesize.js/master/src/filesize.js
//	https://raw.githubusercontent.com/avoidwork/filesize.js/master/src/helpers.js
//	https://raw.githubusercontent.com/avoidwork/filesize.js/master/src/constants.js
//
// Only vectors whose options map onto this port's API (base, round, standard,
// plus the sign/zero/default handling) are reproduced. Upstream options this
// port does not implement (bits, exponent, output, precision, pad, separator,
// spacer, symbols, fullform, locale, roundingMethod) are intentionally omitted;
// see the notes returned by the sync task for those gaps.
//
// Where the upstream test passes a numeric string or BigInt (e.g. "1024",
// " 1024 ", "1.024e3", BigInt(1024)) the numeric value is used here, since this
// port takes a float64 and upstream coerces the argument to a Number with
// identical results.

import "testing"

func TestParityDefault(t *testing.T) {
	// filesize(x) with no options: base auto (decimal), round 2.
	tests := []struct {
		in   float64
		want string
	}{
		{1000, "1 kB"},       // Basic functionality
		{1000000, "1 MB"},    // Basic functionality
		{1000000000, "1 GB"}, // Basic functionality
		{0, "0 B"},           // Basic functionality / zero
		{1, "1 B"},           // Basic functionality (small numbers)
		{512, "512 B"},       // Basic functionality (small numbers)
		{-1000, "-1 kB"},     // Basic functionality (negative)
		{-1000000, "-1 MB"},  // Basic functionality (negative)
		{1024, "1.02 kB"},    // Number input / integer numbers
		{-1024, "-1.02 kB"},  // Number input / integer numbers
		{1536.5, "1.54 kB"},  // Number input / floating point numbers
		{0.5, "1 B"},         // Number input: 0.5 rounds to integer at the B unit
		{1000, "1 kB"},       // String input "1e3"
		{1024, "1.02 kB"},    // String input "1.024e3" / " 1024 "
	}
	for _, tt := range tests {
		if got := FileSize(tt.in); got != tt.want {
			t.Errorf("FileSize(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParityStandards(t *testing.T) {
	tests := []struct {
		in       float64
		standard string
		want     string
	}{
		{1000, "si", "1 kB"},        // Standards: SI
		{1000000, "si", "1 MB"},     // Standards: SI
		{1024, "iec", "1 KiB"},      // Standards: IEC
		{1048576, "iec", "1 MiB"},   // Standards: IEC
		{1024, "jedec", "1 KB"},     // Standards: JEDEC
		{1048576, "jedec", "1 MB"},  // Standards: JEDEC
		{1024, "foobar", "1.02 kB"}, // Unknown standard falls back to decimal default
	}
	for _, tt := range tests {
		got := FileSizeOpts(tt.in, Options{Standard: tt.standard})
		if got != tt.want {
			t.Errorf("FileSizeOpts(%v, standard=%q) = %q, want %q", tt.in, tt.standard, got, tt.want)
		}
	}
}

func TestParityBase(t *testing.T) {
	if got := FileSizeOpts(1024, Options{Base: 2}); got != "1 KiB" {
		t.Errorf("FileSizeOpts(1024, base=2) = %q, want %q", got, "1 KiB")
	}
	if got := FileSizeOpts(1000, Options{Base: 10}); got != "1 kB" {
		t.Errorf("FileSizeOpts(1000, base=10) = %q, want %q", got, "1 kB")
	}
}

func TestParityRounding(t *testing.T) {
	// filesize(1536, {round: n})
	tests := []struct {
		round int
		want  string
	}{
		{1, "1.5 kB"},   // Rounding: round to 1 decimal
		{0, "2 kB"},     // Rounding: round 0 rounds 1.536 -> 2
		{3, "1.536 kB"}, // Rounding: round to 3 decimals
	}
	for _, tt := range tests {
		got := FileSizeOpts(1536, Options{Round: intp(tt.round)})
		if got != tt.want {
			t.Errorf("FileSizeOpts(1536, round=%d) = %q, want %q", tt.round, got, tt.want)
		}
	}
}

func TestParityPartialCombos(t *testing.T) {
	// partial({round: 1, standard: "iec"})(1536) === "1.5 KiB"
	if got := FileSizeOpts(1536, Options{Round: intp(1), Standard: "iec"}); got != "1.5 KiB" {
		t.Errorf("FileSizeOpts(1536, round=1, iec) = %q, want %q", got, "1.5 KiB")
	}
}
