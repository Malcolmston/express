package datefns

import "time"

// This file extends the date-fns port with additional widely-used helpers:
// finer-grained arithmetic (seconds, milliseconds, quarters), more interval
// boundaries (hour, minute, second, quarter, year and ISO week), field getters
// and setters, further difference units, additional same-unit predicates, unix
// timestamp conversions, and interval helpers. All functions are deterministic,
// operate only on their arguments, and preserve each input's time zone unless
// documented otherwise, matching the conventions of the base package.

// AddSeconds returns t advanced by amount seconds (negative moves backward).
func AddSeconds(t time.Time, amount int) time.Time {
	return t.Add(time.Duration(amount) * time.Second)
}

// SubSeconds returns t moved back by amount seconds.
func SubSeconds(t time.Time, amount int) time.Time { return AddSeconds(t, -amount) }

// AddMilliseconds returns t advanced by amount milliseconds.
func AddMilliseconds(t time.Time, amount int) time.Time {
	return t.Add(time.Duration(amount) * time.Millisecond)
}

// SubMilliseconds returns t moved back by amount milliseconds.
func SubMilliseconds(t time.Time, amount int) time.Time { return AddMilliseconds(t, -amount) }

// AddQuarters returns t advanced by amount quarters (three months each).
func AddQuarters(t time.Time, amount int) time.Time { return t.AddDate(0, amount*3, 0) }

// SubQuarters returns t moved back by amount quarters.
func SubQuarters(t time.Time, amount int) time.Time { return AddQuarters(t, -amount) }

// StartOfHour returns t truncated to the beginning of its hour.
func StartOfHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

// EndOfHour returns the last nanosecond of t's hour.
func EndOfHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 59, 59, 999999999, t.Location())
}

// StartOfMinute returns t truncated to the beginning of its minute.
func StartOfMinute(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
}

// EndOfMinute returns the last nanosecond of t's minute.
func EndOfMinute(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, 999999999, t.Location())
}

// StartOfSecond returns t truncated to the beginning of its second.
func StartOfSecond(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}

// EndOfSecond returns the last nanosecond of t's second.
func EndOfSecond(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 999999999, t.Location())
}

// StartOfQuarter returns the first nanosecond of the calendar quarter (Jan, Apr,
// Jul or Oct 1) containing t.
func StartOfQuarter(t time.Time) time.Time {
	q := (int(t.Month()) - 1) / 3
	return time.Date(t.Year(), time.Month(q*3+1), 1, 0, 0, 0, 0, t.Location())
}

// EndOfQuarter returns the last nanosecond of the calendar quarter containing t.
func EndOfQuarter(t time.Time) time.Time {
	start := StartOfQuarter(t)
	return start.AddDate(0, 3, 0).Add(-time.Nanosecond)
}

// EndOfYear returns the last nanosecond of the calendar year containing t.
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), time.December, 31, 23, 59, 59, 999999999, t.Location())
}

// StartOfISOWeek returns the first nanosecond of the ISO-8601 week (Monday) that
// contains t.
func StartOfISOWeek(t time.Time) time.Time {
	d := StartOfDay(t)
	// Go weekday: Sunday=0..Saturday=6; ISO week starts Monday.
	offset := (int(d.Weekday()) + 6) % 7
	return d.AddDate(0, 0, -offset)
}

// EndOfISOWeek returns the last nanosecond of the ISO-8601 week (Sunday) that
// contains t.
func EndOfISOWeek(t time.Time) time.Time {
	start := StartOfISOWeek(t)
	return start.AddDate(0, 0, 7).Add(-time.Nanosecond)
}

// GetHours returns the hour of t in [0,23].
func GetHours(t time.Time) int { return t.Hour() }

// GetMinutes returns the minute of t in [0,59].
func GetMinutes(t time.Time) int { return t.Minute() }

// GetSeconds returns the second of t in [0,59].
func GetSeconds(t time.Time) int { return t.Second() }

// GetMilliseconds returns the millisecond component of t in [0,999].
func GetMilliseconds(t time.Time) int { return t.Nanosecond() / 1e6 }

// GetDate returns the day of the month of t in [1,31] (date-fns getDate).
func GetDate(t time.Time) int { return t.Day() }

// GetMonth returns the zero-indexed month of t in [0,11] (0 = January), matching
// date-fns getMonth.
func GetMonth(t time.Time) int { return int(t.Month()) - 1 }

// GetYear returns the full year of t (date-fns getYear returns the full year in
// modern versions).
func GetYear(t time.Time) int { return t.Year() }

// GetDay returns the day of the week of t in [0,6] (0 = Sunday), matching
// date-fns getDay.
func GetDay(t time.Time) int { return int(t.Weekday()) }

// GetQuarter returns the calendar quarter of t in [1,4].
func GetQuarter(t time.Time) int { return (int(t.Month())-1)/3 + 1 }

// GetISOWeek returns the ISO-8601 week number of t in [1,53].
func GetISOWeek(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

// SetHours returns t with its hour set to h, other fields unchanged.
func SetHours(t time.Time, h int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), h, t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// SetMinutes returns t with its minute set to m.
func SetMinutes(t time.Time, m int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), m, t.Second(), t.Nanosecond(), t.Location())
}

// SetSeconds returns t with its second set to s.
func SetSeconds(t time.Time, s int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), s, t.Nanosecond(), t.Location())
}

// SetDate returns t with its day-of-month set to d, normalizing overflow the
// way time.Date does (day 32 of January becomes February 1).
func SetDate(t time.Time, d int) time.Time {
	return time.Date(t.Year(), t.Month(), d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// SetMonth returns t with its month set to the zero-indexed value m (0 =
// January), matching date-fns setMonth.
func SetMonth(t time.Time, m int) time.Time {
	return time.Date(t.Year(), time.Month(m+1), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// SetYear returns t with its year set to y.
func SetYear(t time.Time, y int) time.Time {
	return time.Date(y, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// LastDayOfMonth returns the first nanosecond of the last day of t's month.
func LastDayOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, -1)
}

// GetDaysInMonth returns the number of days in t's month (28-31).
func GetDaysInMonth(t time.Time) int {
	return StartOfMonth(t).AddDate(0, 1, -1).Day()
}

// DifferenceInMilliseconds returns the signed number of milliseconds a - b.
func DifferenceInMilliseconds(a, b time.Time) int64 {
	return a.Sub(b).Milliseconds()
}

// DifferenceInWeeks returns the signed number of whole weeks between a and b
// (a - b), truncated toward zero.
func DifferenceInWeeks(a, b time.Time) int {
	return DifferenceInDays(a, b) / 7
}

// DifferenceInMonths returns the signed number of whole calendar months between
// a and b (a - b), accounting for the day of month, matching date-fns.
func DifferenceInMonths(a, b time.Time) int {
	sign := 1
	x, y := a, b
	if a.Before(b) {
		x, y = b, a
		sign = -1
	}
	months := (x.Year()-y.Year())*12 + int(x.Month()) - int(y.Month())
	if y.AddDate(0, months, 0).After(x) {
		months--
	}
	return sign * months
}

// DifferenceInQuarters returns the signed number of whole quarters between a and
// b (a - b), truncated toward zero.
func DifferenceInQuarters(a, b time.Time) int {
	return DifferenceInMonths(a, b) / 3
}

// DifferenceInYears returns the signed number of whole years between a and b
// (a - b), accounting for month and day, matching date-fns.
func DifferenceInYears(a, b time.Time) int {
	return DifferenceInMonths(a, b) / 12
}

// IsSameMonth reports whether a and b fall in the same calendar month of the
// same year (compared in a's location).
func IsSameMonth(a, b time.Time) bool {
	bb := b.In(a.Location())
	return a.Year() == bb.Year() && a.Month() == bb.Month()
}

// IsSameYear reports whether a and b fall in the same calendar year (compared in
// a's location).
func IsSameYear(a, b time.Time) bool {
	return a.Year() == b.In(a.Location()).Year()
}

// IsSameHour reports whether a and b fall in the same hour (compared in a's
// location).
func IsSameHour(a, b time.Time) bool {
	return StartOfHour(a).Equal(StartOfHour(b.In(a.Location())))
}

// IsSameMinute reports whether a and b fall in the same minute (compared in a's
// location).
func IsSameMinute(a, b time.Time) bool {
	return StartOfMinute(a).Equal(StartOfMinute(b.In(a.Location())))
}

// IsSameWeek reports whether a and b fall in the same week, using date-fns's
// default Sunday-start week (compared in a's location).
func IsSameWeek(a, b time.Time) bool {
	return StartOfWeek(a).Equal(StartOfWeek(b.In(a.Location())))
}

// IsSameQuarter reports whether a and b fall in the same calendar quarter of the
// same year (compared in a's location).
func IsSameQuarter(a, b time.Time) bool {
	bb := b.In(a.Location())
	return a.Year() == bb.Year() && GetQuarter(a) == GetQuarter(bb)
}

// IsFirstDayOfMonth reports whether t is the first day of its month.
func IsFirstDayOfMonth(t time.Time) bool { return t.Day() == 1 }

// IsLastDayOfMonth reports whether t is the last day of its month.
func IsLastDayOfMonth(t time.Time) bool {
	return t.Day() == GetDaysInMonth(t)
}

// MinDate returns the earliest of the given times. It returns the zero Time when
// called with no arguments.
func MinDate(times ...time.Time) time.Time {
	if len(times) == 0 {
		return time.Time{}
	}
	min := times[0]
	for _, t := range times[1:] {
		if t.Before(min) {
			min = t
		}
	}
	return min
}

// MaxDate returns the latest of the given times. It returns the zero Time when
// called with no arguments.
func MaxDate(times ...time.Time) time.Time {
	if len(times) == 0 {
		return time.Time{}
	}
	max := times[0]
	for _, t := range times[1:] {
		if t.After(max) {
			max = t
		}
	}
	return max
}

// ClosestIndexTo returns the index of the time in candidates nearest to target,
// or -1 when candidates is empty.
func ClosestIndexTo(target time.Time, candidates ...time.Time) int {
	best := -1
	var bestDist time.Duration
	for i, c := range candidates {
		d := target.Sub(c)
		if d < 0 {
			d = -d
		}
		if best == -1 || d < bestDist {
			best = i
			bestDist = d
		}
	}
	return best
}

// ClosestTo returns the time in candidates nearest to target and true, or the
// zero Time and false when candidates is empty.
func ClosestTo(target time.Time, candidates ...time.Time) (time.Time, bool) {
	i := ClosestIndexTo(target, candidates...)
	if i == -1 {
		return time.Time{}, false
	}
	return candidates[i], true
}

// FromUnixTime returns the local Time corresponding to the given Unix timestamp
// in seconds.
func FromUnixTime(sec int64) time.Time { return time.Unix(sec, 0) }

// GetUnixTime returns t as a Unix timestamp in whole seconds.
func GetUnixTime(t time.Time) int64 { return t.Unix() }

// GetTime returns t as a Unix timestamp in milliseconds, matching JavaScript's
// Date.prototype.getTime.
func GetTime(t time.Time) int64 { return t.UnixMilli() }

// ToDate returns the local Time corresponding to a JavaScript-style millisecond
// timestamp.
func ToDate(ms int64) time.Time { return time.UnixMilli(ms) }

// IsWithinInterval reports whether t lies within [start, end] inclusive. If
// start is after end the bounds are swapped.
func IsWithinInterval(t, start, end time.Time) bool {
	if start.After(end) {
		start, end = end, start
	}
	return !t.Before(start) && !t.After(end)
}

// EachDayOfInterval returns the start-of-day time for each day in [start, end]
// inclusive. If start is after end the result is empty.
func EachDayOfInterval(start, end time.Time) []time.Time {
	out := []time.Time{}
	if start.After(end) {
		return out
	}
	d := StartOfDay(start)
	last := StartOfDay(end)
	for !d.After(last) {
		out = append(out, d)
		d = d.AddDate(0, 0, 1)
	}
	return out
}
