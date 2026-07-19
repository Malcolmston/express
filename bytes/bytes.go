// Package bytes converts between byte counts and human readable strings,
// modeled on the npm "bytes" package by TJ Holowaychuk. Units are binary
// (1KB = 1024 bytes). It reimplements that package's two headline operations,
// parsing a size string into a number of bytes and formatting a number of bytes
// into a size string, using only the Go standard library.
//
// This is the utility Express and its middleware use to accept and display byte
// sizes in a form people actually write. You use Parse to turn a configuration
// value or request field such as "100kb" or "1.5mb" into an int64 you can
// compare against, for example when enforcing a request body size limit, and you
// use Format (or FormatOpts) to turn a raw count back into a compact label such
// as "1.5MB" for logs, dashboards, or API responses. Keeping both directions in
// one place means the same unit conventions apply whether a size is being read
// in or printed out.
//
// The unit ladder is binary rather than decimal: B, KB, MB, GB, TB and PB step
// up by factors of 1024, matching the npm package and exposed here as the
// package-level constants B through PB. Parse uses a case-insensitive regular
// expression to accept an optional sign, an integer or decimal number, optional
// spaces, and an optional unit suffix; a value with no unit is interpreted as a
// bare byte count. The parsed number is multiplied by the unit's magnitude and
// floored to an integer, so "1.5kb" yields 1536 and fractional bytes are always
// rounded down toward zero in magnitude. Format walks the ladder to pick the
// largest unit whose magnitude does not exceed the (absolute) value, divides,
// and by default renders up to two decimal places with insignificant trailing
// zeros trimmed, so 1024 becomes "1KB" and 500 becomes "500B".
//
// FormatOpts exposes the knobs the npm options object provides. DecimalPlaces
// overrides the default of two fractional digits, FixedDecimals keeps trailing
// zeros instead of trimming them (so 1GB can render as "1.00GB"), UnitSeparator
// inserts a string such as a space between the number and the unit, and Unit
// forces a specific unit instead of auto-selecting one, letting you express a
// value in KB even when MB would otherwise be chosen. Edge cases are handled
// predictably: zero formats as "0B", negative counts keep their sign while unit
// selection is based on magnitude, and an unrecognized forced unit falls back to
// dividing by one. Parse returns a descriptive error for input that does not
// match the expected shape, including empty strings, non-numeric text, and a
// unit with no number.
//
// The parity with the Node original is close but adapted to Go. Parse and
// Format map to the npm package's parse and format calls, and FormatOptions
// mirrors its options object field-for-field. The differences are idiomatic:
// pointer fields (DecimalPlaces) stand in for JavaScript's "option present or
// undefined" distinction, Parse returns an (int64, error) pair instead of
// returning null on failure, and the numeric type is Go's int64 rather than a
// JavaScript number. The rounding, unit thresholds, and default two-decimal
// trimmed formatting are intended to produce the same strings the npm bytes
// package produces for the same inputs.
package bytes

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Unit magnitudes in bytes (binary).
const (
	// B is one byte.
	B int64 = 1
	// KB is one kibibyte (1024 bytes).
	KB int64 = 1 << 10
	// MB is one mebibyte (1024 KB).
	MB int64 = 1 << 20
	// GB is one gibibyte (1024 MB).
	GB int64 = 1 << 30
	// TB is one tebibyte (1024 GB).
	TB int64 = 1 << 40
	// PB is one pebibyte (1024 TB).
	PB int64 = 1 << 50
)

var unitMap = map[string]int64{
	"b": B, "kb": KB, "mb": MB, "gb": GB, "tb": TB, "pb": PB,
}

var parseRE = regexp.MustCompile(`(?i)^((-|\+)?(\d+(?:\.\d+)?)) *(kb|mb|gb|tb|pb|b)?$`)

// Parse converts a human readable byte string such as "1KB", "1.5MB" or a
// bare number (interpreted as bytes) into an int64 count of bytes. It
// returns an error if the string cannot be parsed.
func Parse(s string) (int64, error) {
	m := parseRE.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, fmt.Errorf("bytes: invalid value %q", s)
	}
	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, fmt.Errorf("bytes: invalid number in %q: %w", s, err)
	}
	unit := "b"
	if m[4] != "" {
		unit = strings.ToLower(m[4])
	}
	return int64(math.Floor(f * float64(unitMap[unit]))), nil
}

// FormatOptions controls how Format renders a byte count.
type FormatOptions struct {
	// DecimalPlaces is the number of fractional digits to show. If nil, 2 is
	// used.
	DecimalPlaces *int
	// FixedDecimals keeps trailing zeros in the fractional part when true.
	FixedDecimals bool
	// UnitSeparator is placed between the number and the unit (default "").
	UnitSeparator string
	// ThousandsSeparator is inserted between every group of three digits in
	// the integer part of the number (e.g. " " renders 1000 bytes as
	// "1 000B"). Empty means no grouping.
	ThousandsSeparator string
	// Unit forces a specific unit (e.g. "MB"). If empty the largest fitting
	// unit is chosen.
	Unit string
}

// Format renders n as a human readable string using the largest fitting
// binary unit, e.g. "1KB", "1.5MB" or "500B".
func Format(n int64) string {
	return FormatOpts(n, FormatOptions{})
}

// FormatOpts renders n as a human readable string using the supplied
// options.
func FormatOpts(n int64, opts FormatOptions) string {
	mag := n
	if mag < 0 {
		mag = -mag
	}

	unit := strings.ToUpper(opts.Unit)
	if unit == "" {
		switch {
		case mag >= PB:
			unit = "PB"
		case mag >= TB:
			unit = "TB"
		case mag >= GB:
			unit = "GB"
		case mag >= MB:
			unit = "MB"
		case mag >= KB:
			unit = "KB"
		default:
			unit = "B"
		}
	}

	div := unitMap[strings.ToLower(unit)]
	if div == 0 {
		div = 1
	}
	val := float64(n) / float64(div)

	decimals := 2
	if opts.DecimalPlaces != nil {
		decimals = *opts.DecimalPlaces
	}

	str := strconv.FormatFloat(val, 'f', decimals, 64)
	if !opts.FixedDecimals {
		str = trimZeros(str)
	}
	if opts.ThousandsSeparator != "" {
		intPart, fracPart, hasFrac := strings.Cut(str, ".")
		str = groupThousands(intPart, opts.ThousandsSeparator)
		if hasFrac {
			str += "." + fracPart
		}
	}
	return str + opts.UnitSeparator + unit
}

// groupThousands inserts sep between every group of three digits in the
// integer string s, counting from the right, leaving any leading sign in
// place. It mirrors the npm bytes package's thousandsSeparator behavior
// (regexp /\B(?=(\d{3})+(?!\d))/g applied to the integer part).
func groupThousands(s, sep string) string {
	sign := ""
	if len(s) > 0 && (s[0] == '-' || s[0] == '+') {
		sign, s = s[:1], s[1:]
	}
	n := len(s)
	if n <= 3 {
		return sign + s
	}
	first := n % 3
	if first == 0 {
		first = 3
	}
	var b strings.Builder
	b.WriteString(s[:first])
	for i := first; i < n; i += 3 {
		b.WriteString(sep)
		b.WriteString(s[i : i+3])
	}
	return sign + b.String()
}

// trimZeros removes trailing fractional zeros (and a dangling decimal point)
// from a fixed-notation number string.
func trimZeros(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
