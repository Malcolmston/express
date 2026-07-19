// Package humanizeduration converts a duration in milliseconds into a human
// readable phrase such as "1 hour" or "1 minute, 1 second". It is a faithful
// port of the npm package "humanize-duration" (English only), reproducing that
// library's default behaviour and the handful of options most commonly reached
// for when formatting durations for end users.
//
// The original JavaScript module exists because raw millisecond counts are
// unfriendly to read: "3661000" means little at a glance, whereas "1 hour, 1
// minute, 1 second" is immediately understood. This port is intended for the
// same job in Go programs, for example rendering an elapsed time, a countdown,
// or a cache TTL in a log line, a CLI, or an HTTP response without pulling in a
// third-party dependency. Only the standard library is used.
//
// Formatting works by decomposing the input into an ordered set of units. Each
// unit has a fixed length in milliseconds: a year is 365.25 days, a month is
// 30.4375 days, a week is 7 days, and days, hours, minutes, seconds and
// milliseconds follow the obvious conversions. The absolute value is divided by
// the largest unit first and the remainder is carried down to the next unit, so
// the counts are computed greedily from largest to smallest. Every unit except
// the last is floored to a whole number; the final unit keeps its fractional
// part, which is why 1500 ms renders as "1.5 seconds". The Largest option caps
// how many non-zero units appear in the output, and Round rolls smaller units
// up into larger ones (with carry) when enabled.
//
// The zero value of Options selects the defaults, which match humanize-duration:
// the unit set y, mo, w, d, h, m, s (milliseconds are only shown when requested
// via Units), a ", " delimiter, a " " spacer between count and name, no cap on
// the number of units, and no rounding. Units that come out to zero are omitted,
// so only the significant parts of a duration are printed. When the whole
// duration reduces to nothing, a single zero-valued phrase built from the
// smallest configured unit is returned, which is why Humanize(0) yields
// "0 seconds" rather than an empty string. Negative inputs are formatted from
// their absolute value with no sign prefix (matching humanize-duration's
// Math.abs handling), and counts are pluralized so that
// exactly 1 uses the singular name and any other value uses the plural.
//
// Parity with Node is close but not total. The port covers English output and
// the Units, Largest, Delimiter, Round and Spacer options; it deliberately
// omits humanize-duration's other locales, custom unit-measure overrides,
// per-unit formatting hooks, decimal/digit-replacement settings and the
// "largest"/"maxDecimalPoints" interactions beyond what is described above.
// Number formatting uses Go's strconv with trailing zeros trimmed, so values
// print without a fixed decimal width. Callers needing behaviour outside this
// subset should supply an explicit Options value or format the count themselves.
package humanizeduration

import (
	"math"
	"strconv"
	"strings"
)

// unitMeasures maps a unit key to its length in milliseconds.
var unitMeasures = map[string]float64{
	"y":  31557600000, // 365.25 days
	"mo": 2629800000,  // 30.4375 days
	"w":  604800000,   // 7 days
	"d":  86400000,
	"h":  3600000,
	"m":  60000,
	"s":  1000,
	"ms": 1,
}

// unitNames maps a unit key to its singular and plural English names.
var unitNames = map[string][2]string{
	"y":  {"year", "years"},
	"mo": {"month", "months"},
	"w":  {"week", "weeks"},
	"d":  {"day", "days"},
	"h":  {"hour", "hours"},
	"m":  {"minute", "minutes"},
	"s":  {"second", "seconds"},
	"ms": {"millisecond", "milliseconds"},
}

// defaultUnits is the ordered list of units used when none are supplied.
var defaultUnits = []string{"y", "mo", "w", "d", "h", "m", "s"}

// Options configures how a duration is rendered.
type Options struct {
	// Units is the ordered list of unit keys to consider. When empty the
	// default (y, mo, w, d, h, m, s) is used. Valid keys are y, mo, w, d, h,
	// m, s and ms.
	Units []string
	// Largest limits the number of units shown. Zero means unlimited.
	Largest int
	// Delimiter separates units. When empty ", " is used.
	Delimiter string
	// Round rounds to the nearest unit, rolling smaller units up.
	Round bool
	// Spacer separates a count from its unit name. When empty " " is used.
	Spacer string
}

// Humanize converts a number of milliseconds into a human readable phrase using
// the default options.
func Humanize(ms int64) string {
	return HumanizeOpts(ms, Options{})
}

type piece struct {
	name  string
	count float64
}

// HumanizeOpts converts a number of milliseconds into a human readable phrase
// honouring the supplied Options.
func HumanizeOpts(ms int64, opts Options) string {
	units := opts.Units
	if len(units) == 0 {
		units = defaultUnits
	}
	delimiter := opts.Delimiter
	if delimiter == "" {
		delimiter = ", "
	}
	spacer := opts.Spacer
	if spacer == "" {
		spacer = " "
	}

	// Upstream (humanize-duration) applies Math.abs to the input, so a negative
	// duration renders identically to its magnitude with no sign prefix.
	value := float64(ms)
	if value < 0 {
		value = -value
	}

	pieces := make([]piece, len(units))
	for i, u := range units {
		um := unitMeasures[u]
		var c float64
		if i == len(units)-1 {
			c = value / um
		} else {
			c = math.Floor(value / um)
		}
		pieces[i] = piece{name: u, count: c}
		value -= c * um
	}

	firstOccupied := 0
	for i := range pieces {
		if pieces[i].count != 0 {
			firstOccupied = i
			break
		}
	}

	if opts.Round {
		for i := len(pieces) - 1; i >= 0; i-- {
			pieces[i].count = math.Round(pieces[i].count)
			if i == 0 {
				break
			}
			ratio := unitMeasures[pieces[i-1].name] / unitMeasures[pieces[i].name]
			if math.Mod(pieces[i].count, ratio) == 0 ||
				(opts.Largest > 0 && opts.Largest-1 < i-firstOccupied) {
				pieces[i-1].count += pieces[i].count / ratio
				pieces[i].count = 0
			}
		}
	}

	var result []piece
	for i := 0; i < len(pieces); i++ {
		if opts.Largest > 0 && len(result) >= opts.Largest {
			break
		}
		if pieces[i].count != 0 {
			result = append(result, pieces[i])
		}
	}

	if len(result) == 0 {
		last := units[len(units)-1]
		return renderPiece(0, last, spacer)
	}

	parts := make([]string, len(result))
	for i, p := range result {
		parts[i] = renderPiece(p.count, p.name, spacer)
	}
	return strings.Join(parts, delimiter)
}

// renderPiece formats a single count/unit pair, pluralizing as needed.
func renderPiece(count float64, unit, spacer string) string {
	names := unitNames[unit]
	name := names[1]
	if count == 1 {
		name = names[0]
	}
	return formatNumber(count) + spacer + name
}

// formatNumber renders a count, dropping any trailing zeros.
func formatNumber(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}
