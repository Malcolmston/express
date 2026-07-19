package prettybytes

import "testing"

// Parity vectors transcribed verbatim from the upstream sindresorhus/pretty-bytes
// test suite:
//
//	https://raw.githubusercontent.com/sindresorhus/pretty-bytes/main/test.js
//
// The reference implementation they encode lives at:
//
//	https://raw.githubusercontent.com/sindresorhus/pretty-bytes/main/index.js
//
// Only vectors expressible through this port's exported API (default SI, Signed,
// Bits, Binary, and MinimumFractionDigits/MaximumFractionDigits) are included.
// Upstream options this port does not implement (locale with a non-default
// language, space:false, nonBreakingSpace, fixedWidth, and BigInt inputs) are
// intentionally omitted; see the task notes for those gaps. Locale vectors whose
// expected output equals the default (locale 'en', false, and undefined) are
// covered by the default cases below.

func pi(i int) *int { return &i }

func TestParityDefault(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0 B"},
		{0.4, "0.4 B"},
		{0.7, "0.7 B"},
		{0.001, "0.001 B"},
		{0.0001, "0.0001 B"},
		{0.1005, "0.101 B"},
		{0.123456, "0.123 B"},
		{10, "10 B"},
		{10.1, "10.1 B"},
		{999, "999 B"},
		{1001, "1 kB"},
		{1e16, "10 PB"},
		{1e30, "1000000 YB"},
		{827181 * 1e26, "82718100 YB"},
	}
	for _, tt := range tests {
		if got := PrettyBytes(tt.in); got != tt.want {
			t.Errorf("PrettyBytes(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParityNegative(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{-0.4, "-0.4 B"},
		{-0.7, "-0.7 B"},
		{-10.1, "-10.1 B"},
		{-999, "-999 B"},
		{-1001, "-1 kB"},
	}
	for _, tt := range tests {
		if got := PrettyBytes(tt.in); got != tt.want {
			t.Errorf("PrettyBytes(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParitySigned(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{42, "+42 B"},
		{-13, "-13 B"},
		{0, " 0 B"},
	}
	for _, tt := range tests {
		if got := PrettyBytesOpts(tt.in, Options{Signed: true}); got != tt.want {
			t.Errorf("PrettyBytesOpts(%v, signed) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParityBits(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0 b"},
		{0.4, "0.4 b"},
		{0.7, "0.7 b"},
		{10, "10 b"},
		{10.1, "10.1 b"},
		{999, "999 b"},
		{1001, "1 kbit"},
		{1e16, "10 Pbit"},
		{1e30, "1000000 Ybit"},
		{827181 * 1e26, "82718100 Ybit"},
	}
	for _, tt := range tests {
		if got := PrettyBytesOpts(tt.in, Options{Bits: true}); got != tt.want {
			t.Errorf("PrettyBytesOpts(%v, bits) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParityBinary(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0 B"},
		{4, "4 B"},
		{10, "10 B"},
		{10.1, "10.1 B"},
		{999, "999 B"},
		{1025, "1 KiB"},
		{1001, "1001 B"},
		{1e16, "8.88 PiB"},
		{1e30, "827181 YiB"},
	}
	for _, tt := range tests {
		if got := PrettyBytesOpts(tt.in, Options{Binary: true}); got != tt.want {
			t.Errorf("PrettyBytesOpts(%v, binary) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParityBitsBinary(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0 b"},
		{4, "4 b"},
		{10, "10 b"},
		{999, "999 b"},
		{1025, "1 kibit"},
		{1e6, "977 kibit"},
		{1e30, "827181 Yibit"},
	}
	for _, tt := range tests {
		if got := PrettyBytesOpts(tt.in, Options{Bits: true, Binary: true}); got != tt.want {
			t.Errorf("PrettyBytesOpts(%v, bits+binary) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParityFractionDigits(t *testing.T) {
	tests := []struct {
		in   float64
		opts Options
		want string
	}{
		{1900, Options{MaximumFractionDigits: pi(1)}, "1.9 kB"},
		{1900, Options{MinimumFractionDigits: pi(3)}, "1.900 kB"},
		{1911, Options{MaximumFractionDigits: pi(1)}, "1.9 kB"},
		{1111, Options{MaximumFractionDigits: pi(2)}, "1.11 kB"},
		{1019, Options{MaximumFractionDigits: pi(3)}, "1.019 kB"},
		{1001, Options{MaximumFractionDigits: pi(3)}, "1.001 kB"},
		{1000, Options{MinimumFractionDigits: pi(1), MaximumFractionDigits: pi(3)}, "1.0 kB"},
		{3942, Options{MinimumFractionDigits: pi(1), MaximumFractionDigits: pi(2)}, "3.94 kB"},
		{59952784, Options{MaximumFractionDigits: pi(1)}, "59.9 MB"},
		{59952784, Options{MinimumFractionDigits: pi(1), MaximumFractionDigits: pi(1)}, "59.9 MB"},
		{4001, Options{MaximumFractionDigits: pi(3), Binary: true}, "3.907 KiB"},
		{18717, Options{MaximumFractionDigits: pi(2), Binary: true}, "18.27 KiB"},
		{18717, Options{MaximumFractionDigits: pi(4), Binary: true}, "18.2783 KiB"},
		{32768, Options{MinimumFractionDigits: pi(2), MaximumFractionDigits: pi(3), Binary: true}, "32.00 KiB"},
		{65536, Options{MinimumFractionDigits: pi(1), MaximumFractionDigits: pi(3), Binary: true}, "64.0 KiB"},
	}
	for _, tt := range tests {
		if got := PrettyBytesOpts(tt.in, tt.opts); got != tt.want {
			t.Errorf("PrettyBytesOpts(%v, %+v) = %q, want %q", tt.in, tt.opts, got, tt.want)
		}
	}
}
