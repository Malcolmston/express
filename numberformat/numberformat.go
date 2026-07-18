// Package numberformat is a standard-library-only Go port of the number
// presentation helpers popularised by the npm packages numeral.js and
// accounting.js, which Express/Node apps use to render numbers for humans:
// thousands grouping, fixed-decimal formatting, currency and percentage
// formatting, ordinal suffixes and compact "1.2k / 3.4M" abbreviation.
//
// The workhorse is FormatFloat, which rounds to a fixed number of decimal
// places (half away from zero), groups the integer part with a configurable
// thousands separator and joins it to the fraction with a configurable decimal
// separator. Comma and CommaFloat are the common English-locale shortcuts
// (comma grouping, dot decimal). FormatCurrency prefixes a symbol and
// FormatPercent multiplies by 100 and appends "%". Ordinal returns the English
// ordinal string for an integer ("1st", "22nd", "113th"), and Abbreviate
// renders large magnitudes compactly with k/M/B/T suffixes. RoundTo exposes the
// underlying decimal rounding.
//
// Negative numbers keep their sign ahead of any grouping, rounding uses
// half-away-from-zero so 2.5 at zero decimals becomes 3, and all functions are
// deterministic and depend only on math, strconv and strings.
package numberformat

import (
	"math"
	"strconv"
	"strings"
)

// RoundTo rounds f to the given number of decimal places using
// half-away-from-zero rounding. A negative decimals value rounds to the left of
// the decimal point (RoundTo(1234, -2) is 1200).
func RoundTo(f float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(f*pow) / pow
}

// group inserts sep between every group of three digits of a pure digit string.
func group(digits, sep string) string {
	if sep == "" || len(digits) <= 3 {
		return digits
	}
	n := len(digits)
	first := n % 3
	var b strings.Builder
	if first > 0 {
		b.WriteString(digits[:first])
	}
	for i := first; i < n; i += 3 {
		if b.Len() > 0 {
			b.WriteString(sep)
		}
		b.WriteString(digits[i : i+3])
	}
	return b.String()
}

// FormatFloat formats f with exactly decimals fractional digits, grouping the
// integer part with thousandsSep and separating the fraction with decSep. The
// value is rounded half away from zero.
func FormatFloat(f float64, decimals int, decSep, thousandsSep string) string {
	if decimals < 0 {
		decimals = 0
	}
	neg := math.Signbit(f) && f != 0
	f = math.Abs(f)
	f = RoundTo(f, decimals)
	fixed := strconv.FormatFloat(f, 'f', decimals, 64)
	intPart, fracPart := fixed, ""
	if i := strings.IndexByte(fixed, '.'); i >= 0 {
		intPart = fixed[:i]
		fracPart = fixed[i+1:]
	}
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	b.WriteString(group(intPart, thousandsSep))
	if decimals > 0 {
		b.WriteString(decSep)
		b.WriteString(fracPart)
	}
	return b.String()
}

// FormatInt formats an integer with the given thousands separator, e.g.
// FormatInt(1234567, ",") is "1,234,567".
func FormatInt(n int64, thousandsSep string) string {
	neg := n < 0
	u := uint64(n)
	if neg {
		u = uint64(-n)
	}
	s := group(strconv.FormatUint(u, 10), thousandsSep)
	if neg {
		return "-" + s
	}
	return s
}

// Comma formats an integer with comma thousands grouping (English locale).
func Comma(n int64) string { return FormatInt(n, ",") }

// CommaFloat formats a float with comma thousands grouping, a dot decimal
// separator and the given number of decimal places.
func CommaFloat(f float64, decimals int) string {
	return FormatFloat(f, decimals, ".", ",")
}

// FormatCurrency formats f as a currency amount, prefixing symbol to the
// comma-grouped, fixed-decimal number. A negative amount places the sign before
// the symbol: FormatCurrency(-1234.5, "$", 2) is "-$1,234.50".
func FormatCurrency(f float64, symbol string, decimals int) string {
	neg := math.Signbit(f) && f != 0
	body := FormatFloat(math.Abs(f), decimals, ".", ",")
	if neg {
		return "-" + symbol + body
	}
	return symbol + body
}

// FormatPercent formats f as a percentage by multiplying by 100 and appending
// "%", with the given number of decimal places: FormatPercent(0.1234, 2) is
// "12.34%".
func FormatPercent(f float64, decimals int) string {
	return FormatFloat(f*100, decimals, ".", "") + "%"
}

// Unformat parses a human-formatted number string back to a float64, ignoring
// any grouping separators, currency symbols, percent signs and surrounding
// text and keeping only digits, a single leading sign and the last decimal
// point. It mirrors accounting.js's unformat, so Unformat("$1,234.56") returns
// 1234.56. It returns an error when no numeric content is found.
func Unformat(s string) (float64, error) {
	var b strings.Builder
	dotAt := strings.LastIndexByte(s, '.')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			b.WriteByte(c)
		case c == '-' && b.Len() == 0:
			b.WriteByte('-')
		case c == '.' && i == dotAt:
			b.WriteByte('.')
		}
	}
	str := b.String()
	if str == "" || str == "-" || str == "." {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseFloat(str, 64)
}

// Ordinal returns the English ordinal representation of n, e.g. 1 -> "1st",
// 2 -> "2nd", 3 -> "3rd", 11 -> "11th", 22 -> "22nd".
func Ordinal(n int) string {
	suffix := "th"
	a := n
	if a < 0 {
		a = -a
	}
	if a%100 < 11 || a%100 > 13 {
		switch a % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return strconv.Itoa(n) + suffix
}

// Abbreviate renders f compactly with a magnitude suffix (k, M, B, T) and the
// given number of decimal places, e.g. Abbreviate(1500, 1) is "1.5k" and
// Abbreviate(2_300_000, 2) is "2.30M". Values below 1000 are formatted plainly.
func Abbreviate(f float64, decimals int) string {
	neg := math.Signbit(f) && f != 0
	abs := math.Abs(f)
	suffixes := []struct {
		threshold float64
		symbol    string
	}{
		{1e12, "T"},
		{1e9, "B"},
		{1e6, "M"},
		{1e3, "k"},
	}
	sign := ""
	if neg {
		sign = "-"
	}
	for _, s := range suffixes {
		if abs >= s.threshold {
			return sign + FormatFloat(abs/s.threshold, decimals, ".", "") + s.symbol
		}
	}
	return sign + FormatFloat(abs, decimals, ".", "")
}
