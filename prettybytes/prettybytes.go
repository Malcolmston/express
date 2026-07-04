// Package prettybytes converts a number of bytes to a human readable string.
//
// It is a faithful port of the npm package "pretty-bytes". By default it uses
// SI units (base 1000): B, kB, MB, GB, TB, PB, EB, ZB, YB and renders numbers
// with up to three significant digits (mirroring JavaScript's toPrecision(3)).
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
		return prefix + formatLocale(number, minFD, maxFD) + separator + units[0]
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
// Number()+toLocaleString: it keeps at most `p` significant digits, strips
// trailing zeros and groups the integer part with commas.
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
	return stripAndGroup(s)
}

// formatLocale rounds to maxFD digits, then removes trailing zeros down to
// minFD and groups the integer part with commas.
func formatLocale(n float64, minFD, maxFD int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatFloat(n, 'f', maxFD, 64)
	intPart := s
	frac := ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart = s[:i]
		frac = s[i+1:]
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

// stripAndGroup strips trailing zeros from the fractional part and groups the
// integer part with commas.
func stripAndGroup(s string) string {
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
	res := group(intPart)
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
