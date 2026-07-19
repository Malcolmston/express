// Package prettybytes converts a number of bytes into a compact, human readable
// string such as "1.34 kB" or "1.5 MiB". It is a faithful port of the npm package
// "pretty-bytes", reproducing its unit tables, rounding and formatting rules using
// only the Go standard library (math, strconv and strings).
//
// Raw byte counts are hard for people to read: "1337000000" conveys far less at a
// glance than "1.34 GB". This package is what you use to render file sizes,
// transfer amounts, memory footprints and similar quantities in logs, CLIs and UIs.
// The default entry point PrettyBytes takes a float64 count and returns the SI
// (base-1000) rendering; PrettyBytesOpts takes the same number plus an Options
// struct for finer control.
//
// By default the package uses SI units (base 1000): B, kB, MB, GB, TB, PB, EB, ZB,
// YB, and formats the mantissa with up to three significant digits, mirroring
// JavaScript's Number.prototype.toPrecision(3) followed by toLocaleString — so
// trailing zeros are stripped ("1 kB", not "1.00 kB"). Matching upstream, the
// default (no fraction-digit options) path does NOT group the integer part with
// commas, so 1e30 renders as "1000000 YB". The correct unit is chosen by taking the
// base-1000 (or base-1024) logarithm of the magnitude and then nudging the
// exponent up or down to correct for floating-point error at the boundaries, so
// values right at a power of the base land on the expected unit.
//
// Options selects alternative renderings. Bits switches to bit units (b, kbit,
// Mbit, ...). Binary switches to base-1024 IEC units (KiB, MiB, ... or kibit,
// Mibit, ... when combined with Bits). Signed prints a leading "+" on positive
// values and a leading space on zero (" 0 B") so columns of signed values align.
// MinimumFractionDigits and MaximumFractionDigits (both *int, nil to leave unset)
// override the default three-significant-digit behaviour with fixed
// fraction-digit bounds resolved the same way Intl.NumberFormat resolves them.
//
// Edge cases and Node parity: zero renders as "0 B" (or " 0 B" when Signed).
// Negative inputs keep their sign and are formatted by magnitude, so -1337 becomes
// "-1.34 kB" regardless of the Signed option. Values below 1 are shown in the base
// unit ("0.4 B"). NaN and the infinities are passed straight through Go's float
// formatter rather than being bucketed into a unit. Magnitudes beyond the largest
// tabulated unit are clamped to that unit (YB / Ybit / YiB / Yibit). These
// behaviours track pretty-bytes; the numeric output is intended to match it for
// the shared inputs.
package prettybytes

import (
	"math"
	"strconv"
	"strings"
)

var byteUnits = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
var bitUnits = []string{"b", "kbit", "Mbit", "Gbit", "Tbit", "Pbit", "Ebit", "Zbit", "Ybit"}
var bibyteUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
var bibitUnits = []string{"b", "kibit", "Mibit", "Gibit", "Tibit", "Pibit", "Eibit", "Zibit", "Yibit"}

// Options configures how a byte count is rendered.
type Options struct {
	// Signed shows a "+" in front of positive numbers (and " " for zero).
	Signed bool
	// Bits uses bit units (b, kbit, ...) instead of byte units.
	Bits bool
	// Binary uses base 1024 units (KiB, MiB, ... or kibit, ...).
	Binary bool
	// MinimumFractionDigits forces a minimum number of fraction digits.
	// When nil, no minimum is imposed.
	MinimumFractionDigits *int
	// MaximumFractionDigits caps the number of fraction digits. When nil the
	// default toPrecision(3) behaviour is used.
	MaximumFractionDigits *int
}

// PrettyBytes converts a number of bytes into a human readable string using
// SI units (base 1000). For example PrettyBytes(1337) returns "1.34 kB".
func PrettyBytes(n float64) string {
	return PrettyBytesOpts(n, Options{})
}

// PrettyBytesOpts converts a number of bytes into a human readable string
// honouring the supplied Options.
func PrettyBytesOpts(number float64, opts Options) string {
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return strconv.FormatFloat(number, 'g', -1, 64)
	}

	units := byteUnits
	switch {
	case opts.Bits && opts.Binary:
		units = bibitUnits
	case opts.Bits:
		units = bitUnits
	case opts.Binary:
		units = bibyteUnits
	}

	const separator = " "

	if opts.Signed && number == 0 {
		return " 0" + separator + units[0]
	}

	isNegative := number < 0
	prefix := ""
	switch {
	case isNegative:
		prefix = "-"
	case opts.Signed:
		prefix = "+"
	}
	if isNegative {
		number = -number
	}

	hasLocale := opts.MinimumFractionDigits != nil || opts.MaximumFractionDigits != nil
	minFD, maxFD := fractionDigits(opts)

	if number < 1 {
		var numStr string
		if hasLocale {
			numStr = formatLocale(number, minFD, maxFD)
		} else {
			numStr = formatPrecision(number, 3)
		}
		return prefix + numStr + separator + units[0]
	}

	base := 1000.0
	var exponent int
	if opts.Binary {
		base = 1024.0
		exponent = int(math.Floor(math.Log(number) / math.Log(1024)))
	} else {
		exponent = int(math.Floor(math.Log10(number) / 3))
	}
	if exponent < 0 {
		exponent = 0
	}
	if exponent > len(units)-1 {
		exponent = len(units) - 1
	}
	// Correct for floating point error in the log estimate.
	for exponent < len(units)-1 && number >= math.Pow(base, float64(exponent+1)) {
		exponent++
	}
	for exponent > 0 && number < math.Pow(base, float64(exponent)) {
		exponent--
	}
	number /= math.Pow(base, float64(exponent))

	var numStr string
	if hasLocale {
		numStr = formatLocale(number, minFD, maxFD)
	} else {
		numStr = formatPrecision(number, 3)
	}

	return prefix + numStr + separator + units[exponent]
}

// fractionDigits resolves the minimum/maximum fraction digits following the
// same defaulting rules as Intl.NumberFormat.
func fractionDigits(opts Options) (minFD, maxFD int) {
	minFD = 0
	if opts.MinimumFractionDigits != nil {
		minFD = *opts.MinimumFractionDigits
	}
	if opts.MaximumFractionDigits != nil {
		maxFD = *opts.MaximumFractionDigits
	} else if opts.MinimumFractionDigits != nil {
		if minFD > 3 {
			maxFD = minFD
		} else {
			maxFD = 3
		}
	} else {
		maxFD = 3
	}
	if maxFD < minFD {
		maxFD = minFD
	}
	return minFD, maxFD
}

// formatPrecision mirrors JavaScript's Number.prototype.toPrecision followed by
// Number()+String(): it keeps at least `p` significant digits (never rounding
// away integer digits) and strips trailing zeros. Upstream calls toLocaleString
// with no locale and no options in this path, which returns the number
// unformatted, so — unlike the locale path — it does NOT group with commas.
func formatPrecision(n float64, p int) string {
	if n == 0 {
		return "0"
	}
	e := int(math.Floor(math.Log10(n)))
	decimals := p - 1 - e
	if decimals < 0 {
		decimals = 0
	}
	s := strconv.FormatFloat(n, 'f', decimals, 64)
	return strip(s)
}

// formatLocale truncates (toward zero) to maxFD digits, then removes trailing
// zeros down to minFD and groups the integer part with commas. Upstream builds
// its Intl.NumberFormat options with `roundingMode: 'trunc'`, so this path must
// truncate rather than round; e.g. 59.952784 at maxFD=1 is "59.9", not "60".
func formatLocale(n float64, minFD, maxFD int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	// Use the shortest round-trip decimal (what Intl.NumberFormat operates on),
	// then cut at maxFD without rounding to mirror roundingMode: 'trunc'. Using
	// the raw double expansion instead would truncate 1.9 (stored as
	// 1.899999...) down to "1.8".
	s := strconv.FormatFloat(n, 'f', -1, 64)
	intPart := s
	frac := ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart = s[:i]
		frac = s[i+1:]
	}
	if len(frac) > maxFD {
		frac = frac[:maxFD]
	}
	for len(frac) < maxFD {
		frac += "0"
	}
	for len(frac) > minFD && strings.HasSuffix(frac, "0") {
		frac = frac[:len(frac)-1]
	}
	res := group(intPart)
	if frac != "" {
		res += "." + frac
	}
	if neg {
		res = "-" + res
	}
	return res
}

// strip removes trailing zeros from the fractional part without grouping the
// integer part, matching upstream's no-locale formatting path.
func strip(s string) string {
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	intPart := s
	frac := ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart = s[:i]
		frac = s[i+1:]
	}
	frac = strings.TrimRight(frac, "0")
	res := intPart
	if frac != "" {
		res += "." + frac
	}
	if neg {
		res = "-" + res
	}
	return res
}

// group inserts comma thousands separators into an integer string.
func group(intPart string) string {
	n := len(intPart)
	if n <= 3 {
		return intPart
	}
	var b strings.Builder
	pre := n % 3
	if pre > 0 {
		b.WriteString(intPart[:pre])
		b.WriteByte(',')
	}
	for i := pre; i < n; i += 3 {
		b.WriteString(intPart[i : i+3])
		if i+3 < n {
			b.WriteByte(',')
		}
	}
	return b.String()
}
