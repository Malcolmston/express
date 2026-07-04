// Package humanizeduration converts a duration in milliseconds into a human
// readable phrase such as "1 hour" or "1 minute, 1 second".
//
// It is a faithful port of the npm package "humanize-duration" (English only).
// The standard unit lengths are used: a year is 365.25 days, a month is
// 30.4375 days and a week is 7 days.
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

	value := float64(ms)
	negative := value < 0
	if negative {
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
	s := strings.Join(parts, delimiter)
	if negative {
		s = "-" + s
	}
	return s
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
