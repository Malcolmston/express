package datefns

import (
	"testing"
	"time"
)

func ref() time.Time {
	// Wed 2021-03-17 08:05:09 UTC
	return time.Date(2021, time.March, 17, 8, 5, 9, 0, time.UTC)
}

func TestTranslateLayout(t *testing.T) {
	got := TranslateLayout("yyyy-MM-dd HH:mm:ss")
	want := "2006-01-02 15:04:05"
	if got != want {
		t.Errorf("TranslateLayout=%q want %q", got, want)
	}
}

func TestFormat(t *testing.T) {
	cases := []struct {
		layout, want string
	}{
		{"yyyy-MM-dd", "2021-03-17"},
		{"HH:mm:ss", "08:05:09"},
		{"EEE MMM dd yyyy", "Wed Mar 17 2021"},
		{"MMMM", "March"},
		{"yy", "21"},
	}
	for _, c := range cases {
		if got := Format(ref(), c.layout); got != c.want {
			t.Errorf("Format(%q)=%q want %q", c.layout, got, c.want)
		}
	}
}

func TestParse(t *testing.T) {
	got, err := Parse("2021-03-17 08:05:09", "yyyy-MM-dd HH:mm:ss")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(ref()) {
		t.Errorf("Parse=%v want %v", got, ref())
	}
}

func TestFormatISO(t *testing.T) {
	if got := FormatISO(ref()); got != "2021-03-17T08:05:09Z" {
		t.Errorf("FormatISO=%q", got)
	}
}

func TestAddSub(t *testing.T) {
	base := ref()
	if got := AddDays(base, 5); got.Day() != 22 {
		t.Errorf("AddDays=%v", got)
	}
	if got := SubDays(base, 17); got.Month() != time.February || got.Day() != 28 {
		t.Errorf("SubDays=%v", got)
	}
	if got := AddWeeks(base, 1); got.Day() != 24 {
		t.Errorf("AddWeeks=%v", got)
	}
	if got := AddMonths(base, 1); got.Month() != time.April {
		t.Errorf("AddMonths=%v", got)
	}
	if got := AddYears(base, 2); got.Year() != 2023 {
		t.Errorf("AddYears=%v", got)
	}
	if got := AddHours(base, 3); got.Hour() != 11 {
		t.Errorf("AddHours=%v", got)
	}
	if got := AddMinutes(base, 10); got.Minute() != 15 {
		t.Errorf("AddMinutes=%v", got)
	}
	if got := SubHours(base, 8); got.Hour() != 0 {
		t.Errorf("SubHours=%v", got)
	}
	if got := SubMinutes(base, 5); got.Minute() != 0 {
		t.Errorf("SubMinutes=%v", got)
	}
	if got := SubWeeks(base, 1); got.Day() != 10 {
		t.Errorf("SubWeeks=%v", got)
	}
	if got := SubMonths(base, 1); got.Month() != time.February {
		t.Errorf("SubMonths=%v", got)
	}
	if got := SubYears(base, 1); got.Year() != 2020 {
		t.Errorf("SubYears=%v", got)
	}
}

func TestDifferences(t *testing.T) {
	a := ref()
	b := a.Add(50 * time.Hour)
	if DifferenceInDays(b, a) != 2 {
		t.Errorf("DifferenceInDays=%d", DifferenceInDays(b, a))
	}
	if DifferenceInHours(b, a) != 50 {
		t.Errorf("DifferenceInHours=%d", DifferenceInHours(b, a))
	}
	if DifferenceInMinutes(b, a) != 3000 {
		t.Errorf("DifferenceInMinutes=%d", DifferenceInMinutes(b, a))
	}
	if DifferenceInSeconds(b, a) != 180000 {
		t.Errorf("DifferenceInSeconds=%d", DifferenceInSeconds(b, a))
	}
}

func TestStartEnd(t *testing.T) {
	base := ref()
	sd := StartOfDay(base)
	if sd.Hour() != 0 || sd.Minute() != 0 || sd.Second() != 0 {
		t.Errorf("StartOfDay=%v", sd)
	}
	ed := EndOfDay(base)
	if ed.Hour() != 23 || ed.Minute() != 59 || ed.Second() != 59 {
		t.Errorf("EndOfDay=%v", ed)
	}
	sm := StartOfMonth(base)
	if sm.Day() != 1 || sm.Hour() != 0 {
		t.Errorf("StartOfMonth=%v", sm)
	}
	em := EndOfMonth(base)
	if em.Day() != 31 || em.Hour() != 23 {
		t.Errorf("EndOfMonth=%v", em)
	}
	// February 2021 has 28 days.
	feb := time.Date(2021, time.February, 10, 0, 0, 0, 0, time.UTC)
	if EndOfMonth(feb).Day() != 28 {
		t.Errorf("EndOfMonth(feb)=%v", EndOfMonth(feb))
	}
	// base is Wednesday; Sunday start of week is 2021-03-14.
	sw := StartOfWeek(base)
	if sw.Weekday() != time.Sunday || sw.Day() != 14 {
		t.Errorf("StartOfWeek=%v", sw)
	}
	ew := EndOfWeek(base)
	if ew.Weekday() != time.Saturday || ew.Day() != 20 {
		t.Errorf("EndOfWeek=%v", ew)
	}
	sy := StartOfYear(base)
	if sy.Month() != time.January || sy.Day() != 1 {
		t.Errorf("StartOfYear=%v", sy)
	}
}

func TestComparisons(t *testing.T) {
	a := ref()
	b := a.Add(time.Hour)
	if !IsBefore(a, b) || IsBefore(b, a) {
		t.Error("IsBefore")
	}
	if !IsAfter(b, a) || IsAfter(a, b) {
		t.Error("IsAfter")
	}
	if !IsEqual(a, a) || IsEqual(a, b) {
		t.Error("IsEqual")
	}
	if !IsSameDay(a, a.Add(3*time.Hour)) {
		t.Error("IsSameDay same")
	}
	if IsSameDay(a, a.Add(48*time.Hour)) {
		t.Error("IsSameDay different")
	}
}

func TestIsWeekend(t *testing.T) {
	sat := time.Date(2021, time.March, 20, 0, 0, 0, 0, time.UTC)
	sun := time.Date(2021, time.March, 21, 0, 0, 0, 0, time.UTC)
	if !IsWeekend(sat) || !IsWeekend(sun) {
		t.Error("weekend days")
	}
	if IsWeekend(ref()) {
		t.Error("Wednesday is not weekend")
	}
}

func TestGetDayOfYear(t *testing.T) {
	if GetDayOfYear(ref()) != 76 {
		t.Errorf("GetDayOfYear=%d", GetDayOfYear(ref()))
	}
}

func TestIsLeapYear(t *testing.T) {
	cases := []struct {
		y    int
		want bool
	}{
		{2020, true},
		{2021, false},
		{1900, false},
		{2000, true},
	}
	for _, c := range cases {
		if got := IsLeapYear(c.y); got != c.want {
			t.Errorf("IsLeapYear(%d)=%v want %v", c.y, got, c.want)
		}
	}
}

func TestFormatDistance(t *testing.T) {
	a := ref()
	cases := []struct {
		d    time.Duration
		want string
	}{
		{10 * time.Second, "less than a minute"},
		{60 * time.Second, "1 minute"},
		{5 * time.Minute, "5 minutes"},
		{60 * time.Minute, "about 1 hour"},
		{5 * time.Hour, "about 5 hours"},
		{24 * time.Hour, "1 day"},
		{3 * 24 * time.Hour, "3 days"},
		{40 * 24 * time.Hour, "about 1 month"},
		{90 * 24 * time.Hour, "3 months"},
		{400 * 24 * time.Hour, "about 1 year"},
		{800 * 24 * time.Hour, "about 2 years"},
	}
	for _, c := range cases {
		if got := FormatDistance(a, a.Add(c.d)); got != c.want {
			t.Errorf("FormatDistance(%v)=%q want %q", c.d, got, c.want)
		}
	}
	// Symmetric / unsigned.
	if FormatDistance(a, a.Add(-3*24*time.Hour)) != "3 days" {
		t.Error("FormatDistance should be unsigned")
	}
}
