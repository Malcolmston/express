# datefns — ports the most commonly used helpers from the npm package date-fns

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/lodash/datefns.svg)](https://pkg.go.dev/github.com/malcolmston/express/lodash/datefns)

Package datefns ports the most commonly used helpers from the npm package
date-fns (https://date-fns.org) to Go's standard time.Time. It covers the four
families that dominate real-world date code: formatting and parsing (Format,
Parse, FormatISO, TranslateLayout), calendar arithmetic (AddDays, AddWeeks,
AddMonths, AddYears, AddHours, AddMinutes and the matching Sub* helpers),
differences (DifferenceInDays/Hours/Minutes/Seconds), interval boundaries
(StartOf/EndOf Day, Week, Month, Year) and a set of predicates and humanizers
(IsBefore, IsAfter, IsSameDay, IsWeekend, IsLeapYear, GetDayOfYear,
FormatDistance). It depends only on the standard library (time, fmt, strings).

The package exists to smooth the port of Express/JavaScript services to Go.
date-fns is the de facto date toolkit in the Node ecosystem, and its function
names and semantics are muscle memory for many developers; reproducing them on
top of time.Time lets that code move across with minimal translation while
staying idiomatic Go. Reach for it when you want date-fns-style call sites
(addDays(d, 5), format(d, "yyyy-MM-dd"), differenceInDays(a, b)) without
hand-writing AddDate offsets and Go reference-layout strings at every use.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/lodash/datefns"
```

## Usage

This is the package's own `ExampleAddDays`, so it compiles and its output is
asserted on every `go test ./datefns/`.

```go
t := time.Date(2021, time.March, 17, 8, 5, 9, 0, time.UTC)
	fmt.Println(datefns.Format(datefns.AddDays(t, 10), "yyyy-MM-dd"))
	fmt.Println(datefns.Format(datefns.AddMonths(t, 1), "yyyy-MM-dd"))
```

```
2021-03-27
2021-04-17
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func AddDays(t time.Time, amount int) time.Time` | AddDays returns t advanced by amount days (amount may be negative). |
| `func AddHours(t time.Time, amount int) time.Time` | AddHours returns t advanced by amount hours. |
| `func AddMilliseconds(t time.Time, amount int) time.Time` | AddMilliseconds returns t advanced by amount milliseconds. |
| `func AddMinutes(t time.Time, amount int) time.Time` | AddMinutes returns t advanced by amount minutes. |
| `func AddMonths(t time.Time, amount int) time.Time` | AddMonths returns t advanced by amount months. |
| `func AddQuarters(t time.Time, amount int) time.Time` | AddQuarters returns t advanced by amount quarters (three months each). |
| `func AddSeconds(t time.Time, amount int) time.Time` | AddSeconds returns t advanced by amount seconds (negative moves backward). |
| `func AddWeeks(t time.Time, amount int) time.Time` | AddWeeks returns t advanced by amount weeks. |
| `func AddYears(t time.Time, amount int) time.Time` | AddYears returns t advanced by amount years. |
| `func ClosestIndexTo(target time.Time, candidates ...time.Time) int` | ClosestIndexTo returns the index of the time in candidates nearest to target, or -1 when candidates is empty. |
| `func ClosestTo(target time.Time, candidates ...time.Time) (time.Time, bool)` | ClosestTo returns the time in candidates nearest to target and true, or the zero Time and false when candidates is empty. |
| `func DifferenceInDays(a, b time.Time) int` | DifferenceInDays returns the number of whole days between the later and earlier of a and b (a minus b), truncated toward zero. |
| `func DifferenceInHours(a, b time.Time) int` | DifferenceInHours returns the number of whole hours in a minus b. |
| `func DifferenceInMilliseconds(a, b time.Time) int64` | DifferenceInMilliseconds returns the signed number of milliseconds a - b. |
| `func DifferenceInMinutes(a, b time.Time) int` | DifferenceInMinutes returns the number of whole minutes in a minus b. |
| `func DifferenceInMonths(a, b time.Time) int` | DifferenceInMonths returns the signed number of whole calendar months between a and b (a - b), accounting for the day of month, matching date-fns. |
| `func DifferenceInQuarters(a, b time.Time) int` | DifferenceInQuarters returns the signed number of whole quarters between a and b (a - b), truncated toward zero. |
| `func DifferenceInSeconds(a, b time.Time) int` | DifferenceInSeconds returns the number of whole seconds in a minus b. |
| `func DifferenceInWeeks(a, b time.Time) int` | DifferenceInWeeks returns the signed number of whole weeks between a and b (a - b), truncated toward zero. |
| `func DifferenceInYears(a, b time.Time) int` | DifferenceInYears returns the signed number of whole years between a and b (a - b), accounting for month and day, matching date-fns. |
| `func EachDayOfInterval(start, end time.Time) []time.Time` | EachDayOfInterval returns the start-of-day time for each day in [start, end] inclusive. |
| `func EndOfDay(t time.Time) time.Time` | EndOfDay returns t set to 23:59:59.999999999 on the same day. |
| `func EndOfHour(t time.Time) time.Time` | EndOfHour returns the last nanosecond of t's hour. |
| `func EndOfISOWeek(t time.Time) time.Time` | EndOfISOWeek returns the last nanosecond of the ISO-8601 week (Sunday) that contains t. |
| `func EndOfMinute(t time.Time) time.Time` | EndOfMinute returns the last nanosecond of t's minute. |
| `func EndOfMonth(t time.Time) time.Time` | EndOfMonth returns the last instant of t's month. |
| `func EndOfQuarter(t time.Time) time.Time` | EndOfQuarter returns the last nanosecond of the calendar quarter containing t. |
| `func EndOfSecond(t time.Time) time.Time` | EndOfSecond returns the last nanosecond of t's second. |
| `func EndOfWeek(t time.Time) time.Time` | EndOfWeek returns the last instant of t's week, with Saturday as the last day. |
| `func EndOfYear(t time.Time) time.Time` | EndOfYear returns the last nanosecond of the calendar year containing t. |
| `func Format(t time.Time, layout string) string` | Format renders t according to a date-fns style layout. |
| `func FormatDistance(a, b time.Time) string` | FormatDistance returns a humanized description of the distance between a and b (for example "3 days" or "about 1 hour"). |
| `func FormatISO(t time.Time) string` | FormatISO renders t in RFC 3339 (ISO 8601) format, e.g. |
| `func FromUnixTime(sec int64) time.Time` | FromUnixTime returns the local Time corresponding to the given Unix timestamp in seconds. |
| `func GetDate(t time.Time) int` | GetDate returns the day of the month of t in [1,31] (date-fns getDate). |
| `func GetDay(t time.Time) int` | GetDay returns the day of the week of t in [0,6] (0 = Sunday), matching date-fns getDay. |
| `func GetDayOfYear(t time.Time) int` | GetDayOfYear returns the ordinal day of the year for t, where January 1 is 1. |
| `func GetDaysInMonth(t time.Time) int` | GetDaysInMonth returns the number of days in t's month (28-31). |
| `func GetHours(t time.Time) int` | GetHours returns the hour of t in [0,23]. |
| `func GetISOWeek(t time.Time) int` | GetISOWeek returns the ISO-8601 week number of t in [1,53]. |
| `func GetMilliseconds(t time.Time) int` | GetMilliseconds returns the millisecond component of t in [0,999]. |
| `func GetMinutes(t time.Time) int` | GetMinutes returns the minute of t in [0,59]. |
| `func GetMonth(t time.Time) int` | GetMonth returns the zero-indexed month of t in [0,11] (0 = January), matching date-fns getMonth. |
| `func GetQuarter(t time.Time) int` | GetQuarter returns the calendar quarter of t in [1,4]. |
| `func GetSeconds(t time.Time) int` | GetSeconds returns the second of t in [0,59]. |
| `func GetTime(t time.Time) int64` | GetTime returns t as a Unix timestamp in milliseconds, matching JavaScript's Date.prototype.getTime. |
| `func GetUnixTime(t time.Time) int64` | GetUnixTime returns t as a Unix timestamp in whole seconds. |
| `func GetYear(t time.Time) int` | GetYear returns the full year of t (date-fns getYear returns the full year in modern versions). |
| `func IsAfter(a, b time.Time) bool` | IsAfter reports whether a occurs strictly after b. |
| `func IsBefore(a, b time.Time) bool` | IsBefore reports whether a occurs strictly before b. |
| `func IsEqual(a, b time.Time) bool` | IsEqual reports whether a and b represent the same instant. |
| `func IsFirstDayOfMonth(t time.Time) bool` | IsFirstDayOfMonth reports whether t is the first day of its month. |
| `func IsLastDayOfMonth(t time.Time) bool` | IsLastDayOfMonth reports whether t is the last day of its month. |
| `func IsLeapYear(year int) bool` | IsLeapYear reports whether the given year is a leap year. |
| `func IsSameDay(a, b time.Time) bool` | IsSameDay reports whether a and b fall on the same calendar day. |
| `func IsSameHour(a, b time.Time) bool` | IsSameHour reports whether a and b fall in the same hour (compared in a's location). |
| `func IsSameMinute(a, b time.Time) bool` | IsSameMinute reports whether a and b fall in the same minute (compared in a's location). |
| `func IsSameMonth(a, b time.Time) bool` | IsSameMonth reports whether a and b fall in the same calendar month of the same year (compared in a's location). |
| `func IsSameQuarter(a, b time.Time) bool` | IsSameQuarter reports whether a and b fall in the same calendar quarter of the same year (compared in a's location). |
| `func IsSameWeek(a, b time.Time) bool` | IsSameWeek reports whether a and b fall in the same week, using date-fns's default Sunday-start week (compared in a's location). |
| `func IsSameYear(a, b time.Time) bool` | IsSameYear reports whether a and b fall in the same calendar year (compared in a's location). |
| `func IsWeekend(t time.Time) bool` | IsWeekend reports whether t falls on a Saturday or Sunday. |
| `func IsWithinInterval(t, start, end time.Time) bool` | IsWithinInterval reports whether t lies within [start, end] inclusive. |
| `func LastDayOfMonth(t time.Time) time.Time` | LastDayOfMonth returns the first nanosecond of the last day of t's month. |
| `func MaxDate(times ...time.Time) time.Time` | MaxDate returns the latest of the given times. |
| `func MinDate(times ...time.Time) time.Time` | MinDate returns the earliest of the given times. |
| `func Parse(value, layout string) (time.Time, error)` | Parse parses value using a date-fns style layout, returning the parsed time. |
| `func SetDate(t time.Time, d int) time.Time` | SetDate returns t with its day-of-month set to d, normalizing overflow the way time.Date does (day 32 of January becomes February 1). |
| `func SetHours(t time.Time, h int) time.Time` | SetHours returns t with its hour set to h, other fields unchanged. |
| `func SetMinutes(t time.Time, m int) time.Time` | SetMinutes returns t with its minute set to m. |
| `func SetMonth(t time.Time, m int) time.Time` | SetMonth returns t with its month set to the zero-indexed value m (0 = January), matching date-fns setMonth. |
| `func SetSeconds(t time.Time, s int) time.Time` | SetSeconds returns t with its second set to s. |
| `func SetYear(t time.Time, y int) time.Time` | SetYear returns t with its year set to y. |
| `func StartOfDay(t time.Time) time.Time` | StartOfDay returns t set to 00:00:00.000000000 on the same day. |
| `func StartOfHour(t time.Time) time.Time` | StartOfHour returns t truncated to the beginning of its hour. |
| `func StartOfISOWeek(t time.Time) time.Time` | StartOfISOWeek returns the first nanosecond of the ISO-8601 week (Monday) that contains t. |
| `func StartOfMinute(t time.Time) time.Time` | StartOfMinute returns t truncated to the beginning of its minute. |
| `func StartOfMonth(t time.Time) time.Time` | StartOfMonth returns the first instant of t's month. |
| `func StartOfQuarter(t time.Time) time.Time` | StartOfQuarter returns the first nanosecond of the calendar quarter (Jan, Apr, Jul or Oct 1) containing t. |
| `func StartOfSecond(t time.Time) time.Time` | StartOfSecond returns t truncated to the beginning of its second. |
| `func StartOfWeek(t time.Time) time.Time` | StartOfWeek returns the first instant of t's week, treating Sunday as the first day of the week (date-fns default). |
| `func StartOfYear(t time.Time) time.Time` | StartOfYear returns the first instant of t's year. |
| `func SubDays(t time.Time, amount int) time.Time` | SubDays returns t moved back by amount days. |
| `func SubHours(t time.Time, amount int) time.Time` | SubHours returns t moved back by amount hours. |
| `func SubMilliseconds(t time.Time, amount int) time.Time` | SubMilliseconds returns t moved back by amount milliseconds. |
| `func SubMinutes(t time.Time, amount int) time.Time` | SubMinutes returns t moved back by amount minutes. |
| `func SubMonths(t time.Time, amount int) time.Time` | SubMonths returns t moved back by amount months. |
| `func SubQuarters(t time.Time, amount int) time.Time` | SubQuarters returns t moved back by amount quarters. |
| `func SubSeconds(t time.Time, amount int) time.Time` | SubSeconds returns t moved back by amount seconds. |
| `func SubWeeks(t time.Time, amount int) time.Time` | SubWeeks returns t moved back by amount weeks. |
| `func SubYears(t time.Time, amount int) time.Time` | SubYears returns t moved back by amount years. |
| `func ToDate(ms int64) time.Time` | ToDate returns the local Time corresponding to a JavaScript-style millisecond timestamp. |
| `func TranslateLayout(layout string) string` | TranslateLayout converts a date-fns style layout string into Go's reference-time layout. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/lodash/datefns).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
