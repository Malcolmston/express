package truncate

import "testing"

func TestTruncateDefault(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		length int
		want   string
	}{
		{"already short", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"basic truncate", "hello world", 8, "hello..."},
		{"the quick brown fox", "The quick brown fox", 10, "The qui..."},
		{"length equals ellipsis", "hello world", 3, "..."},
		{"empty string", "", 5, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.in, tt.length)
			if got != tt.want {
				t.Fatalf("Truncate(%q, %d) = %q, want %q", tt.in, tt.length, got, tt.want)
			}
			if len([]rune(got)) > tt.length && len([]rune(tt.in)) > tt.length {
				t.Fatalf("result %q exceeds length %d", got, tt.length)
			}
		})
	}
}

func TestTruncateWordBoundary(t *testing.T) {
	tests := []struct {
		name string
		in   string
		len  int
		opts Options
		want string
	}{
		{
			name: "word boundary basic",
			in:   "The quick brown fox",
			len:  10,
			opts: Options{Ellipsis: "...", WordBoundary: true},
			want: "The...",
		},
		{
			name: "word boundary keeps whole words",
			in:   "A very long sentence here",
			len:  15,
			opts: Options{Ellipsis: "...", WordBoundary: true},
			want: "A very long...",
		},
		{
			name: "no word boundary cuts mid-word",
			in:   "A very long sentence here",
			len:  15,
			opts: Options{Ellipsis: "...", WordBoundary: false},
			want: "A very long ...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateOpts(tt.in, tt.len, tt.opts)
			if got != tt.want {
				t.Fatalf("TruncateOpts(%q, %d, %+v) = %q, want %q", tt.in, tt.len, tt.opts, got, tt.want)
			}
		})
	}
}

func TestTruncateCustomEllipsis(t *testing.T) {
	tests := []struct {
		name string
		in   string
		len  int
		opts Options
		want string
	}{
		{
			name: "custom ellipsis",
			in:   "hello world",
			len:  8,
			opts: Options{Ellipsis: "…"},
			want: "hello w…",
		},
		{
			name: "no ellipsis",
			in:   "hello world",
			len:  5,
			opts: Options{Ellipsis: ""},
			want: "hello",
		},
		{
			name: "read more ellipsis",
			in:   "hello world foo bar",
			len:  15,
			opts: Options{Ellipsis: " [more]"},
			want: "hello wo [more]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateOpts(tt.in, tt.len, tt.opts)
			if got != tt.want {
				t.Fatalf("TruncateOpts(%q, %d, %+v) = %q, want %q", tt.in, tt.len, tt.opts, got, tt.want)
			}
			if r := []rune(got); len(r) > tt.len {
				t.Fatalf("result %q exceeds length %d", got, tt.len)
			}
		})
	}
}

func TestTruncateUnicode(t *testing.T) {
	// Multi-byte runes must be counted as single characters.
	in := "héllo wörld café"
	got := Truncate(in, 10)
	want := "héllo w..."
	if got != want {
		t.Fatalf("Truncate(%q, 10) = %q, want %q", in, got, want)
	}
	if len([]rune(got)) != 10 {
		t.Fatalf("expected 10 runes, got %d in %q", len([]rune(got)), got)
	}

	// A string of runes already within length is unchanged.
	short := "café"
	if got := Truncate(short, 4); got != short {
		t.Fatalf("Truncate(%q, 4) = %q, want unchanged", short, got)
	}
}
