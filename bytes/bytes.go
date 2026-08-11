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
// package-level constants B through PB. Parse follows the npm package's two-stage
// grammar exactly. First a case-insensitive regular expression accepts an
// optional sign, an integer or decimal number, optional spaces, and one of the
// suffixes kb, mb, gb, tb or pb, anchored at both ends; the number is multiplied
// by the unit's magnitude and floored toward negative infinity (as JavaScript's
// Math.floor does), so "1.5kb" yields 1536 and "-1.5kb" yields -1536. Anything
// that grammar does not match — including a "b" suffix, which the upstream
// alternation omits, a bare number, and any string with surrounding whitespace —
// falls back to JavaScript's parseInt(value, 10) semantics and is interpreted as
// a plain byte count: the longest leading run of decimal digits after optional
// whitespace and a sign wins and the rest of the string is discarded, so
// "1024 please" is 1024, "1zb" is 1, "0x10" is 0 and " 1kb" is 1 rather than
// 1024. Only a string with no leading integer at all, such as "" or "abc", is an
// error. Format walks the ladder to pick the largest unit whose magnitude does
// not exceed the (absolute) value, divides, and by default renders up to two
// decimal places with insignificant trailing zeros trimmed, so 1024 becomes
// "1KB" and 500 becomes "500B".
//
// FormatOpts exposes the knobs the npm options object provides. DecimalPlaces
// overrides the default of two fractional digits, FixedDecimals keeps trailing
// zeros instead of trimming them (so 1GB can render as "1.00GB"), UnitSeparator
// inserts a string such as a space between the number and the unit, and Unit
// forces a specific unit instead of auto-selecting one, letting you express a
// value in KB even when MB would otherwise be chosen. Edge cases are handled
// predictably: zero formats as "0B", negative counts keep their sign while unit
// selection is based on magnitude, and an unrecognized forced unit is ignored in
// favour of auto-selection. Parse returns a descriptive error for input that does
// not match the expected shape, including empty strings, non-numeric text, a
// unit with no number, and values too large to fit in an int64.
//
// The parity with the Node original is close but adapted to Go. Parse and
// Format map to the npm package's parse and format calls, and FormatOptions
// mirrors its options object field-for-field. The differences are idiomatic:
// pointer fields (DecimalPlaces) stand in for JavaScript's "option present or
// undefined" distinction, Parse returns an (int64, error) pair instead of
// returning null on failure, and the numeric type is Go's int64 rather than a
// JavaScript number. Rounding, unit thresholds, trailing-zero trimming and
// thousands grouping reproduce the npm package's output character for character,
// including its quirks: halves round away from zero the way
// Number.prototype.toFixed does, and a fractional part that begins with a zero
// keeps its trailing zeros because upstream's trimming regexp cannot match it.
// The measured comparison against bytes@3.1.2 lives in
// parity/express/nested/bytes.
package bytes

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
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

// parseRE is upstream's grammar character for character:
//
//	/^((-|\+)?(\d+(?:\.\d+)?)) *(kb|mb|gb|tb|pb)$/i
//
// Note what it does not contain: there is no "b" in the alternation, and no
// allowance for whitespace around the value. Anything it misses is handed to
// jsParseInt, exactly as upstream hands it to parseInt(val, 10).
var parseRE = regexp.MustCompile(`(?i)^((-|\+)?(\d+(?:\.\d+)?)) *(kb|mb|gb|tb|pb)$`)

// Parse converts a human readable byte string such as "1KB" or "1.5MB" into an
// int64 count of bytes. A string the unit grammar does not accept — a bare
// number, a "b" suffix, an unknown suffix, anything with surrounding whitespace
// or trailing text — is read as a byte count with JavaScript's
// parseInt(value, 10) semantics, so only input with no leading decimal integer
// at all is an error. See the package comment for the consequences.
func Parse(s string) (int64, error) {
	var f float64
	unit := "b"
	if m := parseRE.FindStringSubmatch(s); m != nil {
		v, err := strconv.ParseFloat(m[1], 64)
		// A literal too large for a float64 becomes ±Inf, which the overflow
		// guard below rejects; upstream's parseFloat produces ±Infinity too.
		if err != nil && !errors.Is(err, strconv.ErrRange) {
			return 0, fmt.Errorf("bytes: invalid number in %q: %w", s, err)
		}
		f, unit = v, strings.ToLower(m[4])
	} else {
		v, ok := jsParseInt(s)
		if !ok {
			return 0, fmt.Errorf("bytes: invalid value %q", s)
		}
		f = v
	}
	n := math.Floor(f * float64(unitMap[unit]))
	// Guard the float64 -> int64 conversion: Go leaves out-of-range conversions
	// implementation-defined, so an input such as "9e9pb" would otherwise yield
	// a meaningless value instead of a diagnosable error. Upstream, having only
	// doubles, returns the oversized number instead.
	if n < math.MinInt64 || n >= math.MaxInt64 {
		return 0, fmt.Errorf("bytes: value %q overflows int64", s)
	}
	return int64(n), nil
}

// jsParseInt implements JavaScript's parseInt(s, 10): skip leading whitespace,
// take an optional sign, consume the longest run of decimal digits and ignore
// whatever follows. The second result is false for what JavaScript would call
// NaN, i.e. no digits were found at all. The digits go through a float64 so a
// long literal loses precision the way upstream's does.
func jsParseInt(s string) (float64, bool) {
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if !isJSWhitespace(r) {
			break
		}
		i += size
	}
	sign := 1.0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		if s[i] == '-' {
			sign = -1
		}
		i++
	}
	start := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == start {
		return 0, false
	}
	f, err := strconv.ParseFloat(s[start:i], 64)
	if err != nil {
		if !errors.Is(err, strconv.ErrRange) {
			return 0, false
		}
		// Only overflow is possible for a run of digits, and parseInt yields
		// Infinity there as well.
		f = math.Inf(1)
	}
	return sign * f, true
}

// isJSWhitespace reports whether r is skipped by parseInt, that is whether it is
// in ECMA-262's WhiteSpace or LineTerminator productions. That is not quite
// Go's unicode.IsSpace: U+0085 is whitespace to Go but not to JavaScript, and
// U+FEFF is whitespace to JavaScript but not to Go.
func isJSWhitespace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x00A0, 0xFEFF, 0x2028, 0x2029:
		return true
	}
	return unicode.Is(unicode.Zs, r)
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
	// Unit forces a specific unit (one of B, KB, MB, GB, TB, PB, matched
	// case-insensitively) and is emitted with the spelling given here. If it is
	// empty or not a recognized unit, the largest fitting unit is chosen
	// instead, matching the npm package.
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
	// Compute the magnitude as a uint64: negating math.MinInt64 overflows back
	// to itself, which would otherwise make it compare as smaller than KB and
	// render as bytes.
	mag := uint64(n)
	if n < 0 {
		mag = -mag
	}

	// Upstream keeps the caller's spelling of a recognized unit and falls back
	// to auto-selection when Unit is empty *or* unrecognized.
	unit := opts.Unit
	if unitMap[strings.ToLower(unit)] == 0 {
		switch {
		case mag >= uint64(PB):
			unit = "PB"
		case mag >= uint64(TB):
			unit = "TB"
		case mag >= uint64(GB):
			unit = "GB"
		case mag >= uint64(MB):
			unit = "MB"
		case mag >= uint64(KB):
			unit = "KB"
		default:
			unit = "B"
		}
	}

	div := unitMap[strings.ToLower(unit)]

	decimals := 2
	if opts.DecimalPlaces != nil {
		decimals = *opts.DecimalPlaces
	}
	// toFixed throws a RangeError outside 0..100. FormatOpts returns a string
	// and has no error channel, so it clamps into range instead; see
	// API-DEVIATIONS.md.
	if decimals < 0 {
		decimals = 0
	}
	if decimals > 100 {
		decimals = 100
	}

	str := jsToFixed(n, div, decimals)
	if !opts.FixedDecimals {
		str = trimDecimals(str)
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

// jsToFixed renders n/div with exactly dp fractional digits, matching
// Number.prototype.toFixed: the sign is split off first and the magnitude is
// then rounded to the nearest representable value, with a tie going to the
// larger one — away from zero, where Go's strconv.FormatFloat would round a tie
// to even and print "1.12" for 1.125 against JavaScript's "1.13". The
// arithmetic is exact (n and div are integers, div a power of two) so no
// intermediate rounding can creep in either.
func jsToFixed(n, div int64, dp int) string {
	sign := ""
	num := new(big.Int).SetInt64(n)
	if n < 0 {
		sign = "-"
		num.Neg(num)
	}
	// Round num/div * 10^dp to an integer, halves upward.
	num.Mul(num, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dp)), nil))
	d := new(big.Int).SetInt64(div)
	rem := new(big.Int)
	num.QuoRem(num, d, rem)
	if rem.Lsh(rem, 1).Cmp(d) >= 0 {
		num.Add(num, big.NewInt(1))
	}
	digits := num.String()
	if dp == 0 {
		return sign + digits
	}
	if len(digits) <= dp {
		digits = strings.Repeat("0", dp-len(digits)+1) + digits
	}
	return sign + digits[:len(digits)-dp] + "." + digits[len(digits)-dp:]
}

// trimDecimalsRE is upstream's trailing-zero trimmer, character for character:
//
//	/(?:\.0*|(\.[^0]+)0+)$/  replaced with  $1
//
// It is deliberately not "strip trailing zeros". A fractional part that begins
// with a zero matches neither branch — the first cannot reach the end of the
// string and the second needs a non-zero digit straight after the point — so
// upstream leaves "1.010" alone, and so does this port.
var trimDecimalsRE = regexp.MustCompile(`(?:\.0*|(\.[^0]+)0+)$`)

// trimDecimals removes insignificant trailing zeros the way upstream does,
// quirk included.
func trimDecimals(s string) string {
	return trimDecimalsRE.ReplaceAllString(s, "$1")
}
