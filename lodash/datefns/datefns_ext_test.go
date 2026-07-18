package datefns

import (
	"testing"
	"time"
)

func extDt(y int, mo time.Month, d, h, mi, s int) time.Time {
	return time.Date(y, mo, d, h, mi, s, 0, time.UTC)
}

func TestExtArithmetic(t *testing.T) {
	base := extDt(2021, time.March, 15, 12, 30, 45)
	if got := AddSeconds(base, 30); got.Second() != 15 || got.Minute() != 31 {
		t.Errorf("AddSeconds = %v", got)
	}
	if got := SubSeconds(base, 45); got.Second() != 0 {
		t.Errorf("SubSeconds = %v", got)
	}
	if got := AddMilliseconds(base, 500); got.Nanosecond() != 500000000 {
		t.Errorf("AddMilliseconds = %v", got)
	}
	if got := AddQuarters(base, 1); got.Month() != time.June {
		t.Errorf("AddQuarters = %v", got)
	}
	if got := SubQuarters(base, 1); got.Month() != time.December || got.Year() != 2020 {
		t.Errorf("SubQuarters = %v", got)
	}
}

func TestBoundaries(t *testing.T) {
	base := extDt(2021, time.March, 15, 12, 30, 45)
	if got := StartOfHour(base); got != extDt(2021, time.March, 15, 12, 0, 0) {
		t.Errorf("StartOfHour = %v", got)
	}
	if got := EndOfHour(base); got.Minute() != 59 || got.Second() != 59 || got.Nanosecond() != 999999999 {
		t.Errorf("EndOfHour = %v", got)
	}
	if got := StartOfMinute(base); got != extDt(2021, time.March, 15, 12, 30, 0) {
		t.Errorf("StartOfMinute = %v", got)
	}
	if got := StartOfQuarter(base); got != extDt(2021, time.January, 1, 0, 0, 0) {
		t.Errorf("StartOfQuarter = %v", got)
	}
	if got := EndOfQuarter(base); got.Month() != time.March || got.Day() != 31 || got.Hour() != 23 {
		t.Errorf("EndOfQuarter = %v", got)
	}
	if got := EndOfYear(base); got.Month() != time.December || got.Day() != 31 {
		t.Errorf("EndOfYear = %v", got)
	}
	// 2021-03-15 is a Monday -> ISO week starts same day
	if got := StartOfISOWeek(base); got != extDt(2021, time.March, 15, 0, 0, 0) {
		t.Errorf("StartOfISOWeek = %v", got)
	}
	// 2021-03-17 is Wednesday -> ISO week start is Monday the 15th
	if got := StartOfISOWeek(extDt(2021, time.March, 17, 9, 0, 0)); got != extDt(2021, time.March, 15, 0, 0, 0) {
		t.Errorf("StartOfISOWeek Wed = %v", got)
	}
}

func TestGetters(t *testing.T) {
	base := time.Date(2021, time.March, 15, 12, 30, 45, 123000000, time.UTC)
	if GetHours(base) != 12 || GetMinutes(base) != 30 || GetSeconds(base) != 45 || GetMilliseconds(base) != 123 {
		t.Error("time getters")
	}
	if GetDate(base) != 15 {
		t.Error("GetDate")
	}
	if GetMonth(base) != 2 { // 0-indexed March
		t.Errorf("GetMonth = %d, want 2", GetMonth(base))
	}
	if GetYear(base) != 2021 {
		t.Error("GetYear")
	}
	if GetDay(base) != 1 { // Monday
		t.Errorf("GetDay = %d, want 1", GetDay(base))
	}
	if GetQuarter(base) != 1 {
		t.Errorf("GetQuarter = %d, want 1", GetQuarter(base))
	}
	if GetISOWeek(base) != 11 {
		t.Errorf("GetISOWeek = %d, want 11", GetISOWeek(base))
	}
}

func TestSetters(t *testing.T) {
	base := extDt(2021, time.March, 15, 12, 30, 45)
	if got := SetHours(base, 9); got.Hour() != 9 {
		t.Error("SetHours")
	}
	if got := SetMinutes(base, 0); got.Minute() != 0 {
		t.Error("SetMinutes")
	}
	if got := SetSeconds(base, 0); got.Second() != 0 {
		t.Error("SetSeconds")
	}
	if got := SetDate(base, 1); got.Day() != 1 {
		t.Error("SetDate")
	}
	if got := SetMonth(base, 0); got.Month() != time.January {
		t.Error("SetMonth")
	}
	if got := SetYear(base, 2000); got.Year() != 2000 {
		t.Error("SetYear")
	}
}

func TestLastDayAndDaysInMonth(t *testing.T) {
	if got := GetDaysInMonth(extDt(2021, time.February, 10, 0, 0, 0)); got != 28 {
		t.Errorf("GetDaysInMonth feb 2021 = %d", got)
	}
	if got := GetDaysInMonth(extDt(2020, time.February, 10, 0, 0, 0)); got != 29 {
		t.Errorf("GetDaysInMonth feb 2020 = %d", got)
	}
	if got := LastDayOfMonth(extDt(2021, time.March, 5, 0, 0, 0)); got.Day() != 31 {
		t.Errorf("LastDayOfMonth = %v", got)
	}
	if !IsFirstDayOfMonth(extDt(2021, time.March, 1, 0, 0, 0)) {
		t.Error("IsFirstDayOfMonth")
	}
	if !IsLastDayOfMonth(extDt(2021, time.March, 31, 0, 0, 0)) {
		t.Error("IsLastDayOfMonth")
	}
}

func TestExtDifferences(t *testing.T) {
	if got := DifferenceInWeeks(extDt(2021, time.March, 15, 0, 0, 0), extDt(2021, time.March, 1, 0, 0, 0)); got != 2 {
		t.Errorf("DifferenceInWeeks = %d", got)
	}
	if got := DifferenceInMonths(extDt(2021, time.March, 1, 0, 0, 0), extDt(2021, time.January, 31, 0, 0, 0)); got != 1 {
		t.Errorf("DifferenceInMonths = %d, want 1", got)
	}
	if got := DifferenceInMonths(extDt(2021, time.January, 31, 0, 0, 0), extDt(2021, time.March, 1, 0, 0, 0)); got != -1 {
		t.Errorf("DifferenceInMonths neg = %d", got)
	}
	if got := DifferenceInYears(extDt(2024, time.March, 15, 0, 0, 0), extDt(2021, time.March, 15, 0, 0, 0)); got != 3 {
		t.Errorf("DifferenceInYears = %d", got)
	}
	if got := DifferenceInQuarters(extDt(2021, time.October, 1, 0, 0, 0), extDt(2021, time.January, 1, 0, 0, 0)); got != 3 {
		t.Errorf("DifferenceInQuarters = %d", got)
	}
	if got := DifferenceInMilliseconds(extDt(2021, time.March, 1, 0, 0, 1), extDt(2021, time.March, 1, 0, 0, 0)); got != 1000 {
		t.Errorf("DifferenceInMilliseconds = %d", got)
	}
}

func TestSamePredicates(t *testing.T) {
	a := extDt(2021, time.March, 15, 12, 0, 0)
	if !IsSameMonth(a, extDt(2021, time.March, 1, 0, 0, 0)) || IsSameMonth(a, extDt(2021, time.April, 1, 0, 0, 0)) {
		t.Error("IsSameMonth")
	}
	if !IsSameYear(a, extDt(2021, time.December, 31, 0, 0, 0)) || IsSameYear(a, extDt(2022, time.January, 1, 0, 0, 0)) {
		t.Error("IsSameYear")
	}
	if !IsSameHour(a, extDt(2021, time.March, 15, 12, 59, 0)) || IsSameHour(a, extDt(2021, time.March, 15, 13, 0, 0)) {
		t.Error("IsSameHour")
	}
	if !IsSameQuarter(a, extDt(2021, time.January, 5, 0, 0, 0)) || IsSameQuarter(a, extDt(2021, time.July, 5, 0, 0, 0)) {
		t.Error("IsSameQuarter")
	}
	// week of Mar 15 2021 (Sun-start): Mar 14 (Sun) .. Mar 20 (Sat)
	if !IsSameWeek(a, extDt(2021, time.March, 14, 0, 0, 0)) || IsSameWeek(a, extDt(2021, time.March, 21, 0, 0, 0)) {
		t.Error("IsSameWeek")
	}
}

func TestMinMaxClosest(t *testing.T) {
	a := extDt(2021, time.January, 1, 0, 0, 0)
	b := extDt(2021, time.June, 1, 0, 0, 0)
	c := extDt(2021, time.December, 1, 0, 0, 0)
	if MinDate(b, a, c) != a {
		t.Error("MinDate")
	}
	if MaxDate(a, c, b) != c {
		t.Error("MaxDate")
	}
	target := extDt(2021, time.November, 15, 0, 0, 0)
	if got, _ := ClosestTo(target, a, b, c); got != c {
		t.Errorf("ClosestTo = %v", got)
	}
	if i := ClosestIndexTo(target, a, b, c); i != 2 {
		t.Errorf("ClosestIndexTo = %d", i)
	}
	if _, ok := ClosestTo(target); ok {
		t.Error("ClosestTo empty")
	}
}

func TestUnixHelpers(t *testing.T) {
	tm := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)
	sec := GetUnixTime(tm)
	if FromUnixTime(sec).UTC() != tm {
		t.Error("Unix round trip")
	}
	if GetTime(tm) != sec*1000 {
		t.Error("GetTime")
	}
	if ToDate(sec*1000).UTC() != tm {
		t.Error("ToDate")
	}
}

func TestIntervals(t *testing.T) {
	start := extDt(2021, time.March, 1, 0, 0, 0)
	end := extDt(2021, time.March, 3, 0, 0, 0)
	if !IsWithinInterval(extDt(2021, time.March, 2, 12, 0, 0), start, end) {
		t.Error("IsWithinInterval inside")
	}
	if IsWithinInterval(extDt(2021, time.March, 4, 0, 0, 0), start, end) {
		t.Error("IsWithinInterval outside")
	}
	days := EachDayOfInterval(start, end)
	if len(days) != 3 || days[0].Day() != 1 || days[2].Day() != 3 {
		t.Errorf("EachDayOfInterval = %v", days)
	}
}

func BenchmarkDifferenceInMonths(b *testing.B) {
	x := extDt(2024, time.March, 1, 0, 0, 0)
	y := extDt(2021, time.January, 15, 0, 0, 0)
	for i := 0; i < b.N; i++ {
		_ = DifferenceInMonths(x, y)
	}
}
