// Package ms converts between time durations and human readable strings,
// modeled on the npm "ms" package. It parses strings such as "2h", "1d" or
// "2.5 days" into time.Durations and formats time.Durations back into short
// ("2h") or long ("2 hours") forms.
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
	msPerMs   = 1.0
	msPerSec  = 1000.0
	msPerMin  = msPerSec * 60
	msPerHour = msPerMin * 60
	msPerDay  = msPerHour * 24
	msPerWeek = msPerDay * 7
	msPerYear = msPerDay * 365.25
)

var parseRE = regexp.MustCompile(`(?i)^(-?(?:\d+)?\.?\d+) *(milliseconds?|msecs?|ms|seconds?|secs?|s|minutes?|mins?|m|hours?|hrs?|h|days?|d|weeks?|w|years?|yrs?|y)?$`)

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
	"y": msPerYear, "yr": msPerYear, "yrs": msPerYear,
	"year": msPerYear, "years": msPerYear,
}

// Format returns the short human readable form of d, choosing the largest
// unit whose magnitude is at least one: days ("d"), hours ("h"), minutes
// ("m"), seconds ("s") or milliseconds ("ms").
func Format(d time.Duration) string {
	msVal := float64(d) / float64(time.Millisecond)
	abs := math.Abs(msVal)
	switch {
	case abs >= msPerDay:
		return roundStr(msVal/msPerDay) + "d"
	case abs >= msPerHour:
		return roundStr(msVal/msPerHour) + "h"
	case abs >= msPerMin:
		return roundStr(msVal/msPerMin) + "m"
	case abs >= msPerSec:
		return roundStr(msVal/msPerSec) + "s"
	default:
		return roundStr(msVal) + "ms"
	}
}

// FormatLong returns the long human readable form of d, such as "2 hours"
// or "1 minute", with correct pluralization.
func FormatLong(d time.Duration) string {
	msVal := float64(d) / float64(time.Millisecond)
	abs := math.Abs(msVal)
	switch {
	case abs >= msPerDay:
		return plural(msVal, abs, msPerDay, "day")
	case abs >= msPerHour:
		return plural(msVal, abs, msPerHour, "hour")
	case abs >= msPerMin:
		return plural(msVal, abs, msPerMin, "minute")
	case abs >= msPerSec:
		return plural(msVal, abs, msPerSec, "second")
	default:
		return roundStr(msVal) + " ms"
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

// roundStr rounds x to the nearest integer (half away from zero) and returns
// its decimal string.
func roundStr(x float64) string {
	return strconv.FormatInt(int64(math.Round(x)), 10)
}
