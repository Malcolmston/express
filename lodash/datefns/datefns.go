// Package datefns ports the most commonly used date-fns helpers to Go's
// time.Time.
//
// Formatting and parsing accept date-fns style tokens (yyyy, MM, dd, HH, mm,
// ss, EEE, MMM, ...) which are translated internally to Go's reference-time
// layout. See Format for the full token table. All functions are deterministic:
// they operate only on the time values passed to them and never read the
// current wall clock.
package datefns

import (
	"fmt"
	"strings"
	"time"
)

// tokenReplacements maps date-fns formatting tokens to the equivalent fragment
// of Go's reference layout (Mon Jan 2 15:04:05 MST 2006). Longer tokens are
// listed before their prefixes so that translation is unambiguous.
var tokenReplacements = []struct{ token, layout string }{
	{"yyyy", "2006"},
	{"yy", "06"},
	{"MMMM", "January"},
	{"MMM", "Jan"},
	{"MM", "01"},
	{"dd", "02"},
	{"EEEE", "Monday"},
	{"EEE", "Mon"},
	{"HH", "15"},
	{"hh", "03"},
	{"mm", "04"},
	{"ss", "05"},
	{"SSS", "000"},
	{"aa", "PM"},
	{"XXX", "-07:00"},
	{"xx", "-0700"},
	{"zzz", "MST"},
}

// TranslateLayout converts a date-fns style layout string into Go's
// reference-time layout. Tokens that are not recognised are passed through
// verbatim, so a layout may freely mix date-fns tokens with literal text.
func TranslateLayout(layout string) string {
	// Placeholder approach: replace each token with a unique sentinel first so
	// that already-substituted layout fragments (e.g. digits) are not matched
	// again by a later, shorter token.
	type repl struct{ sentinel, layout string }
	var repls []repl
	out := layout
	for i, tr := range tokenReplacements {
		sentinel := fmt.Sprintf("\x00%d\x00", i)
		if strings.Contains(out, tr.token) {
			out = strings.ReplaceAll(out, tr.token, sentinel)
			repls = append(repls, repl{sentinel, tr.layout})
		}
	}
	for _, r := range repls {
		out = strings.ReplaceAll(out, r.sentinel, r.layout)
	}
	return out
}

// Format renders t according to a date-fns style layout.
//
// Supported tokens:
//
//	yyyy year (2006)      yy  two-digit year (06)
//	MMMM month name       MMM abbreviated month   MM zero-padded month  dd day
//	EEEE weekday name     EEE abbreviated weekday
//	HH   24-hour          hh  12-hour             mm minute  ss second
//	SSS  milliseconds     aa  AM/PM
//	XXX  zone (+01:00)    xx  zone (+0100)        zzz zone abbrev
//
// Any other characters are emitted literally.
func Format(t time.Time, layout string) string {
	return t.Format(TranslateLayout(layout))
}

// Parse parses value using a date-fns style layout, returning the parsed time.
func Parse(value, layout string) (time.Time, error) {
	return time.Parse(TranslateLayout(layout), value)
}

// FormatISO renders t in RFC 3339 (ISO 8601) format, e.g.
// 2006-01-02T15:04:05Z07:00.
func FormatISO(t time.Time) string {
	return t.Format(time.RFC3339)
}

// AddDays returns t advanced by amount days (amount may be negative).
func AddDays(t time.Time, amount int) time.Time { return t.AddDate(0, 0, amount) }

// AddWeeks returns t advanced by amount weeks.
func AddWeeks(t time.Time, amount int) time.Time { return t.AddDate(0, 0, amount*7) }

// AddMonths returns t advanced by amount months.
func AddMonths(t time.Time, amount int) time.Time { return t.AddDate(0, amount, 0) }

// AddYears returns t advanced by amount years.
func AddYears(t time.Time, amount int) time.Time { return t.AddDate(amount, 0, 0) }

// AddHours returns t advanced by amount hours.
func AddHours(t time.Time, amount int) time.Time {
	return t.Add(time.Duration(amount) * time.Hour)
}

// AddMinutes returns t advanced by amount minutes.
func AddMinutes(t time.Time, amount int) time.Time {
	return t.Add(time.Duration(amount) * time.Minute)
}

// SubDays returns t moved back by amount days.
func SubDays(t time.Time, amount int) time.Time { return AddDays(t, -amount) }

// SubWeeks returns t moved back by amount weeks.
func SubWeeks(t time.Time, amount int) time.Time { return AddWeeks(t, -amount) }

// SubMonths returns t moved back by amount months.
func SubMonths(t time.Time, amount int) time.Time { return AddMonths(t, -amount) }

// SubYears returns t moved back by amount years.
func SubYears(t time.Time, amount int) time.Time { return AddYears(t, -amount) }

// SubHours returns t moved back by amount hours.
func SubHours(t time.Time, amount int) time.Time { return AddHours(t, -amount) }

// SubMinutes returns t moved back by amount minutes.
func SubMinutes(t time.Time, amount int) time.Time { return AddMinutes(t, -amount) }

// DifferenceInDays returns the number of whole days between the later and
// earlier of a and b (a minus b), truncated toward zero.
func DifferenceInDays(a, b time.Time) int {
	return int(a.Sub(b).Hours() / 24)
}

// DifferenceInHours returns the number of whole hours in a minus b.
func DifferenceInHours(a, b time.Time) int {
	return int(a.Sub(b).Hours())
}

// DifferenceInMinutes returns the number of whole minutes in a minus b.
func DifferenceInMinutes(a, b time.Time) int {
	return int(a.Sub(b).Minutes())
}

// DifferenceInSeconds returns the number of whole seconds in a minus b.
func DifferenceInSeconds(a, b time.Time) int {
	return int(a.Sub(b).Seconds())
}

// StartOfDay returns t set to 00:00:00.000000000 on the same day.
func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// EndOfDay returns t set to 23:59:59.999999999 on the same day.
func EndOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// StartOfMonth returns the first instant of t's month.
func StartOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the last instant of t's month.
func EndOfMonth(t time.Time) time.Time {
	return EndOfDay(StartOfMonth(t).AddDate(0, 1, -1))
}

// StartOfWeek returns the first instant of t's week, treating Sunday as the
// first day of the week (date-fns default).
func StartOfWeek(t time.Time) time.Time {
	start := StartOfDay(t)
	return start.AddDate(0, 0, -int(start.Weekday()))
}

// EndOfWeek returns the last instant of t's week, with Saturday as the last day.
func EndOfWeek(t time.Time) time.Time {
	return EndOfDay(StartOfWeek(t).AddDate(0, 0, 6))
}

// StartOfYear returns the first instant of t's year.
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
}

// IsBefore reports whether a occurs strictly before b.
func IsBefore(a, b time.Time) bool { return a.Before(b) }

// IsAfter reports whether a occurs strictly after b.
func IsAfter(a, b time.Time) bool { return a.After(b) }

// IsEqual reports whether a and b represent the same instant.
func IsEqual(a, b time.Time) bool { return a.Equal(b) }

// IsSameDay reports whether a and b fall on the same calendar day. The
// comparison uses a's location for both operands so that days are compared in a
// single, well-defined zone.
func IsSameDay(a, b time.Time) bool {
	b = b.In(a.Location())
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// IsWeekend reports whether t falls on a Saturday or Sunday.
func IsWeekend(t time.Time) bool {
	wd := t.Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

// GetDayOfYear returns the ordinal day of the year for t, where January 1 is 1.
func GetDayOfYear(t time.Time) int { return t.YearDay() }

// IsLeapYear reports whether the given year is a leap year.
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// FormatDistance returns a humanized description of the distance between a and b
// (for example "3 days" or "about 1 hour"). The result is unsigned; use
// IsBefore/IsAfter to determine direction.
func FormatDistance(a, b time.Time) string {
	d := b.Sub(a)
	if d < 0 {
		d = -d
	}
	seconds := int64(d.Seconds())
	switch {
	case seconds < 45:
		return "less than a minute"
	case seconds < 90:
		return "1 minute"
	case seconds < 45*60:
		return fmt.Sprintf("%d minutes", (seconds+30)/60)
	case seconds < 90*60:
		return "about 1 hour"
	case seconds < 24*60*60:
		return fmt.Sprintf("about %d hours", (seconds+30*60)/(60*60))
	case seconds < 42*60*60:
		return "1 day"
	case seconds < 30*24*60*60:
		return fmt.Sprintf("%d days", (seconds+12*60*60)/(24*60*60))
	case seconds < 45*24*60*60:
		return "about 1 month"
	case seconds < 365*24*60*60:
		return fmt.Sprintf("%d months", (seconds+15*24*60*60)/(30*24*60*60))
	case seconds < 545*24*60*60:
		return "about 1 year"
	default:
		return fmt.Sprintf("about %d years", (seconds+182*24*60*60)/(365*24*60*60))
	}
}
