package prettybytes

import "testing"

func intp(i int) *int { return &i }

func TestPrettyBytes(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{10, "10 B"},
		{999, "999 B"},
		{1000, "1 kB"},
		{1337, "1.34 kB"},
		{1500, "1.5 kB"},
		{10000, "10 kB"},
		{100000, "100 kB"},
		{1000000, "1 MB"},
		{1500000, "1.5 MB"},
		{1000000000, "1 GB"},
		{1000000000000, "1 TB"},
		{1000000000000000, "1 PB"},
		{0.4, "0.4 B"},
		{-1337, "-1.34 kB"},
	}
	for _, tt := range tests {
		if got := PrettyBytes(tt.in); got != tt.want {
			t.Errorf("PrettyBytes(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPrettyBytesSigned(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, " 0 B"},
		{1337, "+1.34 kB"},
		{-1337, "-1.34 kB"},
	}
	for _, tt := range tests {
		if got := PrettyBytesOpts(tt.in, Options{Signed: true}); got != tt.want {
			t.Errorf("PrettyBytesOpts(%v, signed) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPrettyBytesBits(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0 b"},
		{1337, "1.34 kbit"},
		{1000000, "1 Mbit"},
	}
	for _, tt := range tests {
		if got := PrettyBytesOpts(tt.in, Options{Bits: true}); got != tt.want {
			t.Errorf("PrettyBytesOpts(%v, bits) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPrettyBytesBinary(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0 B"},
		{1024, "1 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1 MiB"},
	}
	for _, tt := range tests {
		if got := PrettyBytesOpts(tt.in, Options{Binary: true}); got != tt.want {
			t.Errorf("PrettyBytesOpts(%v, binary) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPrettyBytesFractionDigits(t *testing.T) {
	if got := PrettyBytesOpts(1337, Options{MinimumFractionDigits: intp(3)}); got != "1.337 kB" {
		t.Errorf("min fraction digits = %q, want %q", got, "1.337 kB")
	}
	if got := PrettyBytesOpts(1337, Options{MaximumFractionDigits: intp(1)}); got != "1.3 kB" {
		t.Errorf("max fraction digits = %q, want %q", got, "1.3 kB")
	}
}

func TestGrouping(t *testing.T) {
	// 999_999 bytes rounds up to 1000 kB via 3 significant digits. Upstream's
	// default (no locale) path does not group with commas.
	if got := PrettyBytes(999999); got != "1000 kB" {
		t.Errorf("PrettyBytes(999999) = %q, want %q", got, "1000 kB")
	}
}
