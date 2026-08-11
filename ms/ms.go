// Package ms converts between time durations and human readable strings,
// modeled on the npm "ms" package. It parses strings such as "2h", "1d" or
// "2.5 days" into time.Durations and formats time.Durations back into short
// ("2h") or long ("2 hours") forms. The npm original is one of the most widely
// depended-on utilities in the Node ecosystem, used throughout Express and its
// middleware for things like cookie max-ages, cache lifetimes and timeout
// configuration; this package is a standard-library-only Go port of it.
//
// The point of the package is to let humans and configuration files express
// durations in a friendly notation while your code works with a real
// time.Duration. You write "30d" or "2.5 hours" in a config value or a
// command-line flag and Parse turns it into something you can add to a
// time.Time or hand to a timer. Going the other way, Format and FormatLong
// turn a computed duration back into a compact label for logs, UIs or error
// messages. All functions are pure, allocation-light and safe for concurrent
// use.
//
// Parse works in the "string to duration" direction. It understands an
// optional sign, an integer or decimal number, optional surrounding spaces and
// a unit, where the unit may be spelled short (ms, s, m, h, d, w, mo, y) or long
// (millisecond(s), second(s), minute(s), hour(s), day(s), week(s), month(s),
// year(s)) along with common abbreviations like "secs", "mins" and "hrs".
// Matching is case-insensitive. A bare number with no unit is interpreted as
// milliseconds, so Parse("100") is 100 milliseconds, exactly as in the npm
// package. Weeks are 7 days, years are 365.25 days, and a month is a twelfth of
// a year (30.4375 days).
//
// Format and FormatLong work in the "duration to string" direction, which the
// npm ms package selects with its { long } option. Both pick the largest unit
// whose absolute magnitude is at least one — years, months, weeks, days, hours,
// minutes, seconds or milliseconds — and round the count to the nearest whole
// number, except on the milliseconds rung, which is emitted unrounded so that
// half a millisecond prints as "0.5ms". Format produces the terse form ("2h",
// "-1500ms"); FormatLong produces the spelled-out form ("2 hours", "1 minute")
// with pluralization that, matching the original, switches to the plural name
// once the magnitude reaches 1.5 units. Rounding follows JavaScript's
// Math.round, which rounds a half towards positive infinity, so -1500ms formats
// as "-1s". Negative durations are supported in both directions: Parse accepts a
// leading "-" and the formatters preserve the sign.
//
// A few edge cases are worth noting for parity with Node. Empty input is an
// error rather than zero, and an input longer than 100 characters is rejected
// outright as a guard against pathological strings. Anything that does not
// match the number-and-optional-unit grammar (an unknown unit, stray letters,
// multiple numbers) yields an error instead of a silent zero. Unlike the
// JavaScript version, which returns undefined or NaN on bad input, this port
// returns an explicit Go error, so callers can decide how to handle malformed
// values. A duration too large for a time.Duration — beyond roughly 292 years,
// which the unbounded JavaScript number has no trouble with — is likewise an
// error rather than a silently wrapped value.
package ms

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Millisecond magnitudes for each supported unit.
const (
	msPerMs    = 1.0
	msPerSec   = 1000.0
	msPerMin   = msPerSec * 60
	msPerHour  = msPerMin * 60
	msPerDay   = msPerHour * 24
	msPerWeek  = msPerDay * 7
	msPerYear  = msPerDay * 365.25
	msPerMonth = msPerYear / 12
)

var parseRE = regexp.MustCompile(`(?i)^(-?(?:\d+)?\.?\d+) *(milliseconds?|msecs?|ms|seconds?|secs?|s|minutes?|mins?|m|hours?|hrs?|h|days?|d|weeks?|w|months?|mo|years?|yrs?|y)?$`)

// Parse converts a human readable duration string such as "2h", "1d",
// "-1.5h", "100" or "2.5 days" into a time.Duration. A bare number is
// interpreted as milliseconds. It returns an error if the string is empty,
// unreasonably long, or cannot be parsed.
func Parse(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("ms: empty string")
	}
	if len(s) > 100 {
		return 0, fmt.Errorf("ms: string too long")
	}
	m := parseRE.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("ms: invalid duration %q", s)
	}
	val, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, fmt.Errorf("ms: invalid number in %q: %w", s, err)
	}
	unit := "ms"
	if m[2] != "" {
		unit = m[2]
	}
	factor, ok := unitFactors[strings.ToLower(unit)]
	if !ok {
		return 0, fmt.Errorf("ms: unknown unit %q", unit)
	}
	ns := val * factor * float64(time.Millisecond)
	// Guard the float64 -> time.Duration conversion. Upstream returns an
	// unbounded JavaScript number, but a Duration is int64 nanoseconds and tops
	// out near 292 years; an out-of-range conversion is implementation-defined in
	// Go, so report the overflow instead of handing back a wrapped value.
	if math.IsNaN(ns) || ns >= float64(math.MaxInt64) || ns < float64(math.MinInt64) {
		return 0, fmt.Errorf("ms: duration %q is out of range for time.Duration", s)
	}
	return time.Duration(ns), nil
}

// unitFactors maps every accepted unit spelling to its size in milliseconds.
var unitFactors = map[string]float64{
	"ms": msPerMs, "msec": msPerMs, "msecs": msPerMs,
	"millisecond": msPerMs, "milliseconds": msPerMs,
	"s": msPerSec, "sec": msPerSec, "secs": msPerSec,
	"second": msPerSec, "seconds": msPerSec,
	"m": msPerMin, "min": msPerMin, "mins": msPerMin,
	"minute": msPerMin, "minutes": msPerMin,
	"h": msPerHour, "hr": msPerHour, "hrs": msPerHour,
	"hour": msPerHour, "hours": msPerHour,
	"d": msPerDay, "day": msPerDay, "days": msPerDay,
	"w": msPerWeek, "week": msPerWeek, "weeks": msPerWeek,
	"mo": msPerMonth, "month": msPerMonth, "months": msPerMonth,
	"y": msPerYear, "yr": msPerYear, "yrs": msPerYear,
	"year": msPerYear, "years": msPerYear,
}

// Format returns the short human readable form of d, choosing the largest
// unit whose magnitude is at least one: years ("y"), months ("mo"), weeks
// ("w"), days ("d"), hours ("h"), minutes ("m"), seconds ("s") or
// milliseconds ("ms").
func Format(d time.Duration) string {
	msVal := float64(d) / float64(time.Millisecond)
	abs := math.Abs(msVal)
	switch {
	case abs >= msPerYear:
		return roundStr(msVal/msPerYear) + "y"
	case abs >= msPerMonth:
		return roundStr(msVal/msPerMonth) + "mo"
	case abs >= msPerWeek:
		return roundStr(msVal/msPerWeek) + "w"
	case abs >= msPerDay:
		return roundStr(msVal/msPerDay) + "d"
	case abs >= msPerHour:
		return roundStr(msVal/msPerHour) + "h"
	case abs >= msPerMin:
		return roundStr(msVal/msPerMin) + "m"
	case abs >= msPerSec:
		return roundStr(msVal/msPerSec) + "s"
	default:
		// Upstream's milliseconds branch is the one rung of the ladder that does
		// not round: it interpolates the value as-is ("100.5ms").
		return numStr(msVal) + "ms"
	}
}

// FormatLong returns the long human readable form of d, such as "2 hours"
// or "1 minute", with correct pluralization.
func FormatLong(d time.Duration) string {
	msVal := float64(d) / float64(time.Millisecond)
	abs := math.Abs(msVal)
	switch {
	case abs >= msPerYear:
		return plural(msVal, abs, msPerYear, "year")
	case abs >= msPerMonth:
		return plural(msVal, abs, msPerMonth, "month")
	case abs >= msPerWeek:
		return plural(msVal, abs, msPerWeek, "week")
	case abs >= msPerDay:
		return plural(msVal, abs, msPerDay, "day")
	case abs >= msPerHour:
		return plural(msVal, abs, msPerHour, "hour")
	case abs >= msPerMin:
		return plural(msVal, abs, msPerMin, "minute")
	case abs >= msPerSec:
		return plural(msVal, abs, msPerSec, "second")
	default:
		// As in Format: the milliseconds rung is not rounded.
		return numStr(msVal) + " ms"
	}
}

// plural formats a rounded unit count with singular/plural naming, matching
// the ms library's rule of pluralizing when the magnitude is at least 1.5
// units.
func plural(msVal, abs, n float64, name string) string {
	s := roundStr(msVal / n)
	if abs >= n*1.5 {
		return s + " " + name + "s"
	}
	return s + " " + name
}

// roundStr rounds x the way the ms library's JavaScript does and returns its
// decimal string.
//
// JavaScript's Math.round is "half towards +Infinity", not Go's
// math.Round "half away from zero": Math.round(-1.5) is -1 where math.Round
// gives -2. That is visible in ordinary output — formatting -1500ms must yield
// "-1s", not "-2s" — so the rule is reproduced rather than approximated.
func roundStr(x float64) string {
	return strconv.FormatInt(int64(math.Floor(x+0.5)), 10)
}

// numStr renders x the way JavaScript's string interpolation renders a number,
// for the sub-1000 magnitudes the milliseconds rung can hold: shortest decimal
// form that round-trips, no exponent, and no negative zero.
func numStr(x float64) string {
	if x == 0 {
		return "0"
	}
	return strconv.FormatFloat(x, 'f', -1, 64)
}
