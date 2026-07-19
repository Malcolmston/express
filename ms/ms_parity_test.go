package ms

import (
	"testing"
	"time"
)

// Parity vectors transcribed verbatim from the ORIGINAL vercel/ms test suite:
//
//	https://raw.githubusercontent.com/vercel/ms/main/src/index.test.ts
//
// The upstream tests exercise ms() in both directions: ms(string) returns a
// number of milliseconds (here Parse -> time.Duration), and ms(number[, {long}])
// returns a string (here Format / FormatLong). Every expected value below is a
// real value from THAT suite, not invented. Because Go works in time.Duration
// (nanoseconds) rather than a millisecond number, string->number vectors are
// converted with msDur: N upstream milliseconds == time.Duration(N * 1e6 ns).
//
// Upstream unit magnitudes: s=1000, m=60s, h=60m, d=24h, w=7d, y=365.25d,
// mo=y/12 (== 30.4375 days).

// msDur converts an upstream "milliseconds" number into the equivalent Go
// time.Duration, matching how the port represents a parsed value.
func msDur(n float64) time.Duration {
	return time.Duration(n * float64(time.Millisecond))
}

// TestParityParse covers ms(string) -> number from the "ms(string)" and
// "ms(long string)" describe blocks of src/index.test.ts.
func TestParityParse(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		// ms(string)
		{"100", msDur(100)},
		{"1m", msDur(60000)},
		{"1h", msDur(3600000)},
		{"2d", msDur(172800000)},
		{"3w", msDur(1814400000)},
		{"1s", msDur(1000)},
		{"100ms", msDur(100)},
		{"1y", msDur(31557600000)},
		{"1.5h", msDur(5400000)},
		{"1   s", msDur(1000)},   // multiple spaces
		{"1.5H", msDur(5400000)}, // case-insensitive
		{".5ms", msDur(0.5)},     // number starting with "."
		{"-100ms", msDur(-100)},
		{"-1.5h", msDur(-5400000)},
		{"-10.5h", msDur(-37800000)},
		{"-.5h", msDur(-1800000)},
		// ms(long string)
		{"53 milliseconds", msDur(53)},
		{"17 msecs", msDur(17)},
		{"1 sec", msDur(1000)},
		{"1 min", msDur(60000)},
		{"1 hr", msDur(3600000)},
		{"2 days", msDur(172800000)},
		{"1 week", msDur(604800000)},
		{"1 year", msDur(31557600000)},
		{"1.5 hours", msDur(5400000)},
		{"-100 milliseconds", msDur(-100)},
		{"-1.5 hours", msDur(-5400000)},
		{"-.5 hr", msDur(-1800000)},
		// month unit (mo == y/12); parse-supported upstream via the
		// months?|mo alternation in the parse regex.
		{"1mo", msDur(2629800000)},
		{"1 month", msDur(2629800000)},
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

// TestParityParseInvalid covers the "should return NaN if invalid" case and the
// empty-string throw from src/index.test.ts. Upstream returns NaN / throws; the
// Go port returns a non-nil error.
func TestParityParseInvalid(t *testing.T) {
	for _, in := range []string{"☃", "10-.5", "ms", ""} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) expected error", in)
		}
	}
}

// TestParityFormatShort covers ms(number) -> short string from the "ms(number)"
// describe block of src/index.test.ts.
func TestParityFormatShort(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{500, "500ms"},
		{-500, "-500ms"},
		{1000, "1s"},
		{10000, "10s"},
		{-1000, "-1s"},
		{-10000, "-10s"},
		{60 * 1000, "1m"},
		{60 * 10000, "10m"},
		{-1 * 60 * 1000, "-1m"},
		{-1 * 60 * 10000, "-10m"},
		{60 * 60 * 1000, "1h"},
		{60 * 60 * 10000, "10h"},
		{-1 * 60 * 60 * 1000, "-1h"},
		{-1 * 60 * 60 * 10000, "-10h"},
		{24 * 60 * 60 * 1000, "1d"},
		{24 * 60 * 60 * 6000, "6d"},
		{-1 * 24 * 60 * 60 * 1000, "-1d"},
		{-1 * 24 * 60 * 60 * 6000, "-6d"},
		{1 * 7 * 24 * 60 * 60 * 1000, "1w"},
		{2 * 7 * 24 * 60 * 60 * 1000, "2w"},
		{-1 * 1 * 7 * 24 * 60 * 60 * 1000, "-1w"},
		{-1 * 2 * 7 * 24 * 60 * 60 * 1000, "-2w"},
		{30.4375 * 24 * 60 * 60 * 1000, "1mo"},
		{30.4375 * 24 * 60 * 60 * 1200, "1mo"},
		{30.4375 * 24 * 60 * 60 * 10000, "10mo"},
		{-1 * 30.4375 * 24 * 60 * 60 * 1000, "-1mo"},
		{-1 * 30.4375 * 24 * 60 * 60 * 10000, "-10mo"},
		{365.25*24*60*60*1000 + 1, "1y"},
		{365.25*24*60*60*1200 + 1, "1y"},
		{365.25*24*60*60*10000 + 1, "10y"},
		{-1*365.25*24*60*60*1000 - 1, "-1y"},
		{-1*365.25*24*60*60*10000 - 1, "-10y"},
		{234234234, "3d"},   // should round
		{-234234234, "-3d"}, // should round
	}
	for _, c := range cases {
		if got := Format(msDur(c.in)); got != c.want {
			t.Errorf("Format(%vms) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestParityFormatLong covers ms(number, {long:true}) -> long string from the
// "ms(number, { long: true })" describe block of src/index.test.ts.
func TestParityFormatLong(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{500, "500 ms"},
		{-500, "-500 ms"},
		{1000, "1 second"},
		{1200, "1 second"},
		{10000, "10 seconds"},
		{-1000, "-1 second"},
		{-1200, "-1 second"},
		{-10000, "-10 seconds"},
		{60 * 1000, "1 minute"},
		{60 * 1200, "1 minute"},
		{60 * 10000, "10 minutes"},
		{-1 * 60 * 1000, "-1 minute"},
		{-1 * 60 * 1200, "-1 minute"},
		{-1 * 60 * 10000, "-10 minutes"},
		{60 * 60 * 1000, "1 hour"},
		{60 * 60 * 1200, "1 hour"},
		{60 * 60 * 10000, "10 hours"},
		{-1 * 60 * 60 * 1000, "-1 hour"},
		{-1 * 60 * 60 * 1200, "-1 hour"},
		{-1 * 60 * 60 * 10000, "-10 hours"},
		{1 * 24 * 60 * 60 * 1000, "1 day"},
		{1 * 24 * 60 * 60 * 1200, "1 day"},
		{6 * 24 * 60 * 60 * 1000, "6 days"},
		{-1 * 1 * 24 * 60 * 60 * 1000, "-1 day"},
		{-1 * 1 * 24 * 60 * 60 * 1200, "-1 day"},
		{-1 * 6 * 24 * 60 * 60 * 1000, "-6 days"},
		{1 * 7 * 24 * 60 * 60 * 1000, "1 week"},
		{2 * 7 * 24 * 60 * 60 * 1000, "2 weeks"},
		{-1 * 1 * 7 * 24 * 60 * 60 * 1000, "-1 week"},
		{-1 * 2 * 7 * 24 * 60 * 60 * 1000, "-2 weeks"},
		{30.4375 * 24 * 60 * 60 * 1000, "1 month"},
		{30.4375 * 24 * 60 * 60 * 1200, "1 month"},
		{30.4375 * 24 * 60 * 60 * 10000, "10 months"},
		{-1 * 30.4375 * 24 * 60 * 60 * 1000, "-1 month"},
		{-1 * 30.4375 * 24 * 60 * 60 * 1200, "-1 month"},
		{-1 * 30.4375 * 24 * 60 * 60 * 10000, "-10 months"},
		{365.25*24*60*60*1000 + 1, "1 year"},
		{365.25*24*60*60*1200 + 1, "1 year"},
		{365.25*24*60*60*10000 + 1, "10 years"},
		{-1*365.25*24*60*60*1000 - 1, "-1 year"},
		{-1*365.25*24*60*60*1200 - 1, "-1 year"},
		{-1*365.25*24*60*60*10000 - 1, "-10 years"},
		{234234234, "3 days"},   // should round
		{-234234234, "-3 days"}, // should round
	}
	for _, c := range cases {
		if got := FormatLong(msDur(c.in)); got != c.want {
			t.Errorf("FormatLong(%vms) = %q, want %q", c.in, got, c.want)
		}
	}
}
