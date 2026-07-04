package humanizeduration

import "testing"

func TestHumanize(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 seconds"},
		{1000, "1 second"},
		{2000, "2 seconds"},
		{60000, "1 minute"},
		{61000, "1 minute, 1 second"},
		{3600000, "1 hour"},
		{3661000, "1 hour, 1 minute, 1 second"},
		{86400000, "1 day"},
		{172800000, "2 days"},
		{604800000, "1 week"},
		{2629800000, "1 month"},
		{31557600000, "1 year"},
		{31557600000 + 86400000, "1 year, 1 day"},
		{90000, "1 minute, 30 seconds"},
		{1500, "1.5 seconds"},
	}
	for _, tt := range tests {
		if got := Humanize(tt.in); got != tt.want {
			t.Errorf("Humanize(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestHumanizeLargest(t *testing.T) {
	// 1 hour, 1 minute, 1 second -> largest 1 -> "1 hour"
	if got := HumanizeOpts(3661000, Options{Largest: 1}); got != "1 hour" {
		t.Errorf("largest 1 = %q, want %q", got, "1 hour")
	}
	if got := HumanizeOpts(3661000, Options{Largest: 2}); got != "1 hour, 1 minute" {
		t.Errorf("largest 2 = %q, want %q", got, "1 hour, 1 minute")
	}
}

func TestHumanizeUnits(t *testing.T) {
	// Use ms unit explicitly.
	if got := HumanizeOpts(1500, Options{Units: []string{"s", "ms"}}); got != "1 second, 500 milliseconds" {
		t.Errorf("units s,ms = %q, want %q", got, "1 second, 500 milliseconds")
	}
}

func TestHumanizeDelimiter(t *testing.T) {
	if got := HumanizeOpts(61000, Options{Delimiter: " and "}); got != "1 minute and 1 second" {
		t.Errorf("delimiter = %q, want %q", got, "1 minute and 1 second")
	}
}

func TestHumanizeRound(t *testing.T) {
	// 1500 ms rounds to 2 seconds.
	if got := HumanizeOpts(1500, Options{Round: true}); got != "2 seconds" {
		t.Errorf("round = %q, want %q", got, "2 seconds")
	}
	// 90000 ms (1m30s) rounds seconds up.
	if got := HumanizeOpts(89000, Options{Round: true}); got != "1 minute, 29 seconds" {
		t.Errorf("round = %q, want %q", got, "1 minute, 29 seconds")
	}
}

func TestHumanizeNegative(t *testing.T) {
	if got := Humanize(-61000); got != "-1 minute, 1 second" {
		t.Errorf("Humanize(-61000) = %q, want %q", got, "-1 minute, 1 second")
	}
}
