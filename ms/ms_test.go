package ms

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"100", 100 * time.Millisecond},
		{"1m", time.Minute},
		{"1min", time.Minute},
		{"10s", 10 * time.Second},
		{"2h", 2 * time.Hour},
		{"1d", 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"-1.5h", -90 * time.Minute},
		{"2.5 days", time.Duration(2.5 * float64(24*time.Hour))},
		{"1y", time.Duration(365.25 * float64(24*time.Hour))},
		{"1ms", time.Millisecond},
		{".5s", 500 * time.Millisecond},
		{"1 hour", time.Hour},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("Parse(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "abc", "1 fortnight", "12..3s"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) expected error", in)
		}
	}
}

func TestFormat(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{2 * time.Hour, "2h"},
		{time.Minute, "1m"},
		{-3 * 24 * time.Hour, "-3d"},
		{500 * time.Millisecond, "500ms"},
		{10 * time.Second, "10s"},
		{24 * time.Hour, "1d"},
	}
	for _, c := range cases {
		if got := Format(c.in); got != c.want {
			t.Errorf("Format(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatLong(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{2 * time.Hour, "2 hours"},
		{time.Minute, "1 minute"},
		{24 * time.Hour, "1 day"},
		{2 * 24 * time.Hour, "2 days"},
		{10 * time.Second, "10 seconds"},
		{time.Second, "1 second"},
	}
	for _, c := range cases {
		if got := FormatLong(c.in); got != c.want {
			t.Errorf("FormatLong(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	d := 2 * time.Hour
	s := Format(d)
	got, err := Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	if got != d {
		t.Fatalf("round-trip: %v -> %q -> %v", d, s, got)
	}
}
