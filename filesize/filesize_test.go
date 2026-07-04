package filesize

import "testing"

func intp(i int) *int { return &i }

func TestFileSizeBase10(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1000, "1 kB"},
		{1337, "1.34 kB"},
		{1500, "1.5 kB"},
		{1000000, "1 MB"},
		{1500000, "1.5 MB"},
		{1000000000, "1 GB"},
		{1000000000000, "1 TB"},
		{-1337, "-1.34 kB"},
	}
	for _, tt := range tests {
		if got := FileSize(tt.in); got != tt.want {
			t.Errorf("FileSize(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFileSizeBase2(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1 MiB"},
		{1073741824, "1 GiB"},
		{1337, "1.31 KiB"},
	}
	for _, tt := range tests {
		if got := FileSizeOpts(tt.in, Options{Base: 2}); got != tt.want {
			t.Errorf("FileSizeOpts(%v, base2) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFileSizeJedec(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{1024, "1 KB"},
		{1048576, "1 MB"},
		{1536, "1.5 KB"},
	}
	for _, tt := range tests {
		if got := FileSizeOpts(tt.in, Options{Base: 2, Standard: "jedec"}); got != tt.want {
			t.Errorf("FileSizeOpts(%v, jedec) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFileSizeRound(t *testing.T) {
	if got := FileSizeOpts(1337, Options{Round: intp(0)}); got != "1 kB" {
		t.Errorf("round 0 = %q, want %q", got, "1 kB")
	}
	if got := FileSizeOpts(1337, Options{Round: intp(3)}); got != "1.337 kB" {
		t.Errorf("round 3 = %q, want %q", got, "1.337 kB")
	}
}
