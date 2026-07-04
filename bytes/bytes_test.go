package bytes

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"1kb", 1024},
		{"1KB", 1024},
		{"1mb", 1048576},
		{"1gb", 1 << 30},
		{"1tb", 1 << 40},
		{"1pb", 1 << 50},
		{"1.5MB", int64(1.5 * float64(1<<20))},
		{"1024", 1024},
		{"1b", 1},
		{"-1kb", -1024},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("Parse(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "abc", "1 foo", "kb"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) expected error", in)
		}
	}
}

func TestFormat(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{1024, "1KB"},
		{int64(1.5 * float64(1<<20)), "1.5MB"},
		{1 << 30, "1GB"},
		{500, "500B"},
		{0, "0B"},
		{1 << 40, "1TB"},
		{1 << 50, "1PB"},
	}
	for _, c := range cases {
		if got := Format(c.in); got != c.want {
			t.Errorf("Format(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatOpts(t *testing.T) {
	dp := 2
	if got := FormatOpts(1<<30, FormatOptions{DecimalPlaces: &dp, FixedDecimals: true}); got != "1.00GB" {
		t.Errorf("got %q, want 1.00GB", got)
	}
	if got := FormatOpts(1024, FormatOptions{UnitSeparator: " "}); got != "1 KB" {
		t.Errorf("got %q, want '1 KB'", got)
	}
	if got := FormatOpts(1<<20, FormatOptions{Unit: "KB"}); got != "1024KB" {
		t.Errorf("got %q, want 1024KB", got)
	}
}

func TestRoundTrip(t *testing.T) {
	for _, n := range []int64{1024, 1 << 20, 1 << 30, 512} {
		s := Format(n)
		got, err := Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q): %v", s, err)
		}
		if got != n {
			t.Errorf("round-trip %d -> %q -> %d", n, s, got)
		}
	}
}
