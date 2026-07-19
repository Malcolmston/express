// Package filesize converts a number of bytes into a human readable string,
// a faithful port of the npm package "filesize".
//
// Reach for this package whenever a raw byte count needs to be shown to a
// person: file managers, upload progress, download sizes, disk usage, and log
// output all read better as "1.34 kB" than as "1337". FileSize covers the
// common case with sensible defaults, while FileSizeOpts exposes the knobs
// needed to match a particular convention.
//
// The rendering works by choosing an exponent e such that the value falls in
// the range [1, base) when divided by base^e, then formatting that quotient.
// The base is 1000 for SI/decimal units and 1024 for binary units. Because the
// exponent is first estimated with a floating-point logarithm, the code then
// nudges it up or down against exact powers of the divisor to correct for
// rounding error at the unit boundaries, guaranteeing that a value like exactly
// 1000 renders as "1 kB" rather than "1000 B".
//
// Three unit families are supported through the Standard option: "si" gives the
// decimal units B, kB, MB, GB, ...; "iec" gives the binary units B, KiB, MiB,
// GiB, ...; and "jedec" gives the hybrid B, KB, MB, GB, ... with 1024-based
// magnitudes but SI-style spelling. When Standard is empty it is derived from
// Base, so base 2 defaults to "iec" and base 10 defaults to "si". The number of
// fractional digits is controlled by Round (default 2), and trailing zeros
// together with a dangling decimal point are stripped, so "1.50" becomes "1.5"
// and "1.00" becomes "1".
//
// Edge cases follow the Node original closely. A value of zero always renders as
// "0 B" regardless of base or standard. Negative inputs are formatted by their
// magnitude with a leading "-", e.g. "-1.34 kB". Values large enough to exceed
// the last known unit (YB or YiB) are clamped to that final unit rather than
// overflowing the table. The parity gap is that this port returns only the
// formatted string: the npm library's ability to return an array or a rich
// object, to force a fixed exponent, or to customize the symbol table and
// separators is intentionally omitted in favour of a single string result.
package filesize

import (
	"math"
	"strconv"
	"strings"
)

var siUnits = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
var iecUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
var jedecUnits = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

// Options configures how a byte count is rendered.
type Options struct {
	// Base is the numeric base: 10 (divide by 1000) or 2 (divide by 1024).
	// Zero means the default of 10.
	Base int
	// Round is the number of decimal places to keep. When nil the default of
	// 2 is used.
	Round *int
	// Standard selects the unit family: "si" (kB), "iec" (KiB) or "jedec"
	// (KB). When empty it is derived from Base: base 2 uses "iec" and base 10
	// uses "si".
	Standard string
}

// FileSize converts a number of bytes into a human readable string using the
// default options (base 10, two decimals). For example FileSize(1337) returns
// "1.34 kB".
func FileSize(bytes float64) string {
	return FileSizeOpts(bytes, Options{})
}

// FileSizeOpts converts a number of bytes into a human readable string honouring
// the supplied Options.
func FileSizeOpts(bytes float64, opts Options) string {
	base := opts.Base
	if base == 0 {
		base = 10
	}
	round := 2
	if opts.Round != nil {
		round = *opts.Round
	}
	standard := opts.Standard
	if standard == "" {
		if base == 2 {
			standard = "iec"
		} else {
			standard = "si"
		}
	}

	var divisor float64
	var units []string
	switch standard {
	case "iec":
		divisor, units = 1024, iecUnits
	case "jedec":
		divisor, units = 1024, jedecUnits
	default: // "si"
		divisor, units = 1000, siUnits
	}

	negative := bytes < 0
	if negative {
		bytes = -bytes
	}
	prefix := ""
	if negative {
		prefix = "-"
	}

	if bytes == 0 {
		return prefix + "0 " + units[0]
	}

	exponent := int(math.Floor(math.Log(bytes) / math.Log(divisor)))
	if exponent < 0 {
		exponent = 0
	}
	if exponent > len(units)-1 {
		exponent = len(units) - 1
	}
	// Correct for floating point error in the log estimate.
	for exponent < len(units)-1 && bytes >= math.Pow(divisor, float64(exponent+1)) {
		exponent++
	}
	for exponent > 0 && bytes < math.Pow(divisor, float64(exponent)) {
		exponent--
	}

	val := bytes / math.Pow(divisor, float64(exponent))

	// Mirror upstream's applyRounding: decimal places are only applied above the
	// byte unit, so the B unit (exponent 0) always renders as an integer. This
	// matches the Node original, e.g. filesize(0.5) == "1 B".
	decimals := 0
	p := 1.0
	if exponent > 0 && round > 0 {
		decimals = round
		p = math.Pow(10, float64(round))
	}
	var r float64
	if p == 1 {
		r = math.Round(val)
	} else {
		r = math.Round(val*p) / p
	}
	// Auto-increment when rounding overflows the unit's ceiling, so a value like
	// 999999 renders as "1 MB" rather than "1000 kB".
	if r == divisor && exponent < len(units)-1 {
		r = 1
		decimals = 0
		exponent++
	}

	s := stripTrailingZeros(strconv.FormatFloat(r, 'f', decimals, 64))
	return prefix + s + " " + units[exponent]
}

// stripTrailingZeros removes trailing zeros (and a dangling decimal point).
func stripTrailingZeros(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
