package humanizeduration

import "testing"

// Parity tests derived from the upstream npm library "humanize-duration"
// (EvanHahn/HumanizeDuration.js). Every input -> expected-output vector below is
// copied verbatim from the upstream project's own test suite; none are invented.
//
// Sources (fetched 2026-07-19):
//   https://raw.githubusercontent.com/EvanHahn/HumanizeDuration.js/main/test/definitions/en.tsv
//   https://raw.githubusercontent.com/EvanHahn/HumanizeDuration.js/main/test/humanizer.js
//   https://raw.githubusercontent.com/EvanHahn/HumanizeDuration.js/main/humanize-duration.js
//
// The en.tsv vectors are exercised by upstream test/languages.js with the fixed
// options { delimiter: "+", units: ["y","mo","w","d","h","m","s","ms"] }. The
// humanizer.js vectors use the library defaults except where an option is noted.
//
// Vectors that require features the Go port does not expose (custom
// unitMeasures, decimal / maxDecimalPoints / digitReplacements, conjunction /
// serialComma, non-English languages) or that require sub-millisecond fractional
// inputs (the Go API takes int64) are intentionally omitted.

// enUnits mirrors the upstream en.tsv option set.
var enUnits = []string{"y", "mo", "w", "d", "h", "m", "s", "ms"}

// TestParityEnDefinitions replays the upstream test/definitions/en.tsv table.
func TestParityEnDefinitions(t *testing.T) {
	opts := Options{Delimiter: "+", Units: enUnits}
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 milliseconds"},
		{1, "1 millisecond"},
		{2, "2 milliseconds"},
		{5, "5 milliseconds"},
		{12, "12 milliseconds"},
		{420, "420 milliseconds"},
		{500, "500 milliseconds"},
		{1000, "1 second"},
		{1500, "1 second+500 milliseconds"},
		{2000, "2 seconds"},
		{2500, "2 seconds+500 milliseconds"},
		{3000, "3 seconds"},
		{1001, "1 second+1 millisecond"},
		{1002, "1 second+2 milliseconds"},
		{2001, "2 seconds+1 millisecond"},
		{2003, "2 seconds+3 milliseconds"},
		{1200, "1 second+200 milliseconds"},
		{6900, "6 seconds+900 milliseconds"},
		{30000, "30 seconds"},
		{60000, "1 minute"},
		{90000, "1 minute+30 seconds"},
		{120000, "2 minutes"},
		{150000, "2 minutes+30 seconds"},
		{180000, "3 minutes"},
		{61000, "1 minute+1 second"},
		{78000, "1 minute+18 seconds"},
		{61001, "1 minute+1 second+1 millisecond"},
		{61005, "1 minute+1 second+5 milliseconds"},
		{62001, "1 minute+2 seconds+1 millisecond"},
		{62005, "1 minute+2 seconds+5 milliseconds"},
		{121001, "2 minutes+1 second+1 millisecond"},
		{121007, "2 minutes+1 second+7 milliseconds"},
		{138001, "2 minutes+18 seconds+1 millisecond"},
		{138006, "2 minutes+18 seconds+6 milliseconds"},
		{1800000, "30 minutes"},
		{3600000, "1 hour"},
		{5400000, "1 hour+30 minutes"},
		{7200000, "2 hours"},
		{9000000, "2 hours+30 minutes"},
		{10800000, "3 hours"},
		{3660000, "1 hour+1 minute"},
		{3720000, "1 hour+2 minutes"},
		{10860000, "3 hours+1 minute"},
		{11040000, "3 hours+4 minutes"},
		{43200000, "12 hours"},
		{86400000, "1 day"},
		{129600000, "1 day+12 hours"},
		{172800000, "2 days"},
		{216000000, "2 days+12 hours"},
		{259200000, "3 days"},
		{302400000, "3 days+12 hours"},
		{604800000, "1 week"},
		{907200000, "1 week+3 days+12 hours"},
		{1209600000, "2 weeks"},
		{1512000000, "2 weeks+3 days+12 hours"},
		{1814400000, "3 weeks"},
		{1314900000, "2 weeks+1 day+5 hours+15 minutes"},
		{2629800000, "1 month"},
		{3944700000, "1 month+2 weeks+1 day+5 hours+15 minutes"},
		{5259600000, "2 months"},
		{6574500000, "2 months+2 weeks+1 day+5 hours+15 minutes"},
		{7889400000, "3 months"},
		{15778800000, "6 months"},
		{31557600000, "1 year"},
		{47336400000, "1 year+6 months"},
		{63115200000, "2 years"},
		{78894000000, "2 years+6 months"},
		{94672800000, "3 years"},
		// Upstream applies Math.abs to the input, so a negative renders the
		// same as its magnitude with no sign prefix.
		{-420, "420 milliseconds"},
	}
	for _, tt := range tests {
		if got := HumanizeOpts(tt.in, opts); got != tt.want {
			t.Errorf("HumanizeOpts(%d, en.tsv opts) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestParityDelimiter mirrors humanizer.js "can change the delimiter".
func TestParityDelimiter(t *testing.T) {
	opts := Options{Delimiter: "+"}
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 seconds"},
		{1000, "1 second"},
		{363000, "6 minutes+3 seconds"},
	}
	for _, tt := range tests {
		if got := HumanizeOpts(tt.in, opts); got != tt.want {
			t.Errorf("HumanizeOpts(%d, delimiter +) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestParitySpacer mirrors humanizer.js "can change the spacer".
func TestParitySpacer(t *testing.T) {
	opts := Options{Spacer: " whole "}
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 whole seconds"},
		{1000, "1 whole second"},
		{260040000, "3 whole days, 14 whole minutes"},
	}
	for _, tt := range tests {
		if got := HumanizeOpts(tt.in, opts); got != tt.want {
			t.Errorf("HumanizeOpts(%d, spacer) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestParityUnits mirrors humanizer.js "can change the units" (units: ["d"]).
func TestParityUnits(t *testing.T) {
	opts := Options{Units: []string{"d"}}
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 days"},
		{21600000, "0.25 days"}, // ms("6h")
		{604800000, "7 days"},   // ms("7d")
	}
	for _, tt := range tests {
		if got := HumanizeOpts(tt.in, opts); got != tt.want {
			t.Errorf("HumanizeOpts(%d, units [d]) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestParityRound mirrors humanizer.js "can do simple rounding".
func TestParityRound(t *testing.T) {
	opts := Options{Round: true}
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 seconds"},
		{499, "0 seconds"},
		{500, "1 second"},
		{1000, "1 second"},
		{1499, "1 second"},
		{1500, "2 seconds"},
		{121499, "2 minutes, 1 second"},
		{121500, "2 minutes, 2 seconds"},
	}
	for _, tt := range tests {
		if got := HumanizeOpts(tt.in, opts); got != tt.want {
			t.Errorf("HumanizeOpts(%d, round) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestParityRoundUnits mirrors humanizer.js 'can do rounding with the "units" option'.
func TestParityRoundUnits(t *testing.T) {
	tests := []struct {
		in    int64
		units []string
		want  string
	}{
		{86364000, []string{"y", "mo", "w", "d", "h"}, "1 day"},
		{1209564000, []string{"y", "mo", "w", "d", "h"}, "2 weeks"},
		{3692131200000, []string{"y", "mo"}, "117 years"},
		{3692131200001, []string{"y", "mo", "w", "d", "h", "m"}, "116 years, 11 months, 4 weeks, 1 day, 4 hours, 30 minutes"},
	}
	for _, tt := range tests {
		got := HumanizeOpts(tt.in, Options{Round: true, Units: tt.units})
		if got != tt.want {
			t.Errorf("HumanizeOpts(%d, round+units) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestParityRoundLargest mirrors humanizer.js 'can do rounding with the "largest" option'.
func TestParityRoundLargest(t *testing.T) {
	tests := []struct {
		in      int64
		largest int
		want    string
	}{
		{3692131200000, 1, "117 years"},
		{3692131200000, 2, "117 years"},
		{3692131200001, 100, "116 years, 11 months, 4 weeks, 1 day, 4 hours, 30 minutes"},
		{2838550, 3, "47 minutes, 19 seconds"},
	}
	for _, tt := range tests {
		got := HumanizeOpts(tt.in, Options{Round: true, Largest: tt.largest})
		if got != tt.want {
			t.Errorf("HumanizeOpts(%d, round+largest %d) = %q, want %q", tt.in, tt.largest, got, tt.want)
		}
	}
}

// TestParityLargest mirrors humanizer.js "can ask for the largest units".
func TestParityLargest(t *testing.T) {
	tests := []struct {
		in      int64
		largest int
		want    string
	}{
		{0, 2, "0 seconds"},
		{1000, 2, "1 second"},
		{2000, 2, "2 seconds"},
		{540360012, 2, "6 days, 6 hours"},
		{540360012, 3, "6 days, 6 hours, 6 minutes"},
		{540360012, 100, "6 days, 6 hours, 6 minutes, 0.012 seconds"},
	}
	for _, tt := range tests {
		got := HumanizeOpts(tt.in, Options{Largest: tt.largest})
		if got != tt.want {
			t.Errorf("HumanizeOpts(%d, largest %d) = %q, want %q", tt.in, tt.largest, got, tt.want)
		}
	}
}
