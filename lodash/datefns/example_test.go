package datefns_test

import (
	"fmt"
	"time"

	"github.com/malcolmston/express/lodash/datefns"
)

// This example renders a fixed instant with Format using date-fns style tokens.
// The layout mixes tokens (yyyy, MM, dd, HH, mm, ss) with literal separators,
// and Format translates it into Go's reference-time layout before formatting.
// A fixed time.Time in UTC is used so the output is fully deterministic and
// independent of the machine's clock or zone. Unrecognized characters such as the
// dashes and colon are emitted verbatim. This is the direct analogue of date-fns's
// format(date, "yyyy-MM-dd HH:mm:ss").
func ExampleFormat() {
	t := time.Date(2021, time.March, 17, 8, 5, 9, 0, time.UTC)
	fmt.Println(datefns.Format(t, "yyyy-MM-dd HH:mm:ss"))
	fmt.Println(datefns.Format(t, "EEE MMM dd yyyy"))
	// Output:
	// 2021-03-17 08:05:09
	// Wed Mar 17 2021
}

// This example advances a date by whole calendar units with AddDays and AddMonths.
// Both helpers delegate to time.Time.AddDate, so month arithmetic normalizes
// overflowing dates just as Go does. Starting from a fixed UTC instant, adding ten
// days crosses into the same month while adding one month moves to April on the
// same day-of-month. Negative amounts (via the Sub* helpers) would move backward
// instead. A fixed input keeps the output deterministic.
func ExampleAddDays() {
	t := time.Date(2021, time.March, 17, 8, 5, 9, 0, time.UTC)
	fmt.Println(datefns.Format(datefns.AddDays(t, 10), "yyyy-MM-dd"))
	fmt.Println(datefns.Format(datefns.AddMonths(t, 1), "yyyy-MM-dd"))
	// Output:
	// 2021-03-27
	// 2021-04-17
}

// This example measures the whole-day gap between two instants with
// DifferenceInDays. The function computes the first argument minus the second and
// truncates toward zero, so it counts only fully elapsed 24-hour spans. Here the
// two fixed UTC dates are two days and twelve hours apart, which truncates down to
// two whole days. Reversing the arguments would yield a negative result. Using
// fixed UTC times avoids any daylight-saving ambiguity in the count.
func ExampleDifferenceInDays() {
	a := time.Date(2021, time.March, 20, 20, 0, 0, 0, time.UTC)
	b := time.Date(2021, time.March, 18, 8, 0, 0, 0, time.UTC)
	fmt.Println(datefns.DifferenceInDays(a, b))
	fmt.Println(datefns.DifferenceInDays(b, a))
	// Output:
	// 2
	// -2
}

// This example snaps a timestamp to interval boundaries with StartOfMonth and
// EndOfMonth. StartOfMonth returns the first instant of the month at midnight,
// while EndOfMonth returns the last representable nanosecond of the month's final
// day. Both preserve the input's location, so passing a UTC time keeps the result
// in UTC. March 2021 has 31 days, so the end boundary lands on the 31st at
// 23:59:59. The instants are formatted with an ISO-like layout including
// milliseconds to show the truncation to the last nanosecond.
func ExampleStartOfMonth() {
	t := time.Date(2021, time.March, 17, 8, 5, 9, 0, time.UTC)
	fmt.Println(datefns.Format(datefns.StartOfMonth(t), "yyyy-MM-dd HH:mm:ss"))
	fmt.Println(datefns.Format(datefns.EndOfMonth(t), "yyyy-MM-dd HH:mm:ss"))
	// Output:
	// 2021-03-01 00:00:00
	// 2021-03-31 23:59:59
}

// This example computes the bounds of the week containing a fixed date with
// StartOfWeek and EndOfWeek. Following date-fns's default, Sunday is treated as
// the first day of the week and Saturday as the last. March 17, 2021 is a
// Wednesday, so the surrounding week runs from Sunday the 14th to Saturday the
// 20th. Both helpers preserve the input's UTC location, keeping the output stable.
// The weekday token EEE confirms the boundaries fall on Sunday and Saturday.
func ExampleStartOfWeek() {
	t := time.Date(2021, time.March, 17, 8, 5, 9, 0, time.UTC)
	fmt.Println(datefns.Format(datefns.StartOfWeek(t), "EEE yyyy-MM-dd"))
	fmt.Println(datefns.Format(datefns.EndOfWeek(t), "EEE yyyy-MM-dd"))
	// Output:
	// Sun 2021-03-14
	// Sat 2021-03-20
}

// This example produces a humanized, unsigned distance between two instants with
// FormatDistance. The function ignores direction and rounds the elapsed time to a
// coarse bucket using date-fns's thresholds, so a gap of a little over an hour
// reads as "about 1 hour" and three days reads as "3 days". Because the result is
// unsigned, swapping the arguments does not change it; use IsBefore or IsAfter to
// recover direction. Fixed UTC inputs make the buckets deterministic.
func ExampleFormatDistance() {
	a := time.Date(2021, time.March, 17, 8, 0, 0, 0, time.UTC)
	fmt.Println(datefns.FormatDistance(a, a.Add(65*time.Minute)))
	fmt.Println(datefns.FormatDistance(a, a.Add(3*24*time.Hour)))
	// Output:
	// about 1 hour
	// 3 days
}
