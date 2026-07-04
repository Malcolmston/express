// Package bytes converts between byte counts and human readable strings,
// modeled on the npm "bytes" package by TJ Holowaychuk. Units are binary
// (1KB = 1024 bytes).
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
	B  int64 = 1
	KB int64 = 1 << 10
	MB int64 = 1 << 20
	GB int64 = 1 << 30
	TB int64 = 1 << 40
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
	return str + opts.UnitSeparator + unit
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
