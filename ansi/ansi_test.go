package ansi

import "testing"

func TestConvenienceWrappers(t *testing.T) {
	tests := []struct {
		got  string
		want string
	}{
		{Bold("x"), "\x1b[1mx\x1b[0m"},
		{Dim("x"), "\x1b[2mx\x1b[0m"},
		{Italic("x"), "\x1b[3mx\x1b[0m"},
		{Underline("x"), "\x1b[4mx\x1b[0m"},
		{Inverse("x"), "\x1b[7mx\x1b[0m"},
		{Strikethrough("x"), "\x1b[9mx\x1b[0m"},
		{Red("x"), "\x1b[31mx\x1b[0m"},
		{Green("x"), "\x1b[32mx\x1b[0m"},
		{Blue("x"), "\x1b[34mx\x1b[0m"},
		{Yellow("x"), "\x1b[33mx\x1b[0m"},
		{Black("x"), "\x1b[30mx\x1b[0m"},
		{Magenta("x"), "\x1b[35mx\x1b[0m"},
		{Cyan("x"), "\x1b[36mx\x1b[0m"},
		{White("x"), "\x1b[37mx\x1b[0m"},
		{Gray("x"), "\x1b[90mx\x1b[0m"},
		{BgRed("x"), "\x1b[41mx\x1b[0m"},
		{BgGreen("x"), "\x1b[42mx\x1b[0m"},
		{BgBlue("x"), "\x1b[44mx\x1b[0m"},
		{BgYellow("x"), "\x1b[43mx\x1b[0m"},
	}
	for i, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("case %d: got %q, want %q", i, tt.got, tt.want)
		}
	}
}

func TestStyleCompose(t *testing.T) {
	got := New().Bold().Red().Apply("x")
	want := "\x1b[1;31mx\x1b[0m"
	if got != want {
		t.Errorf("compose = %q, want %q", got, want)
	}
	// immutability: base is not mutated by extension
	base := New().Bold()
	_ = base.Red()
	if got := base.Apply("y"); got != "\x1b[1my\x1b[0m" {
		t.Errorf("immutability broken: %q", got)
	}
}

func TestEmptyStyle(t *testing.T) {
	if got := New().Apply("plain"); got != "plain" {
		t.Errorf("empty style = %q", got)
	}
}

func TestRGBAndHex(t *testing.T) {
	if got := RGB(255, 0, 0, "x"); got != "\x1b[38;2;255;0;0mx\x1b[0m" {
		t.Errorf("RGB = %q", got)
	}
	if got := BgRGB(0, 128, 255, "x"); got != "\x1b[48;2;0;128;255mx\x1b[0m" {
		t.Errorf("BgRGB = %q", got)
	}
	if got := Hex("#ff0000", "x"); got != "\x1b[38;2;255;0;0mx\x1b[0m" {
		t.Errorf("Hex = %q", got)
	}
	if got := Hex("#f00", "x"); got != "\x1b[38;2;255;0;0mx\x1b[0m" {
		t.Errorf("Hex short = %q", got)
	}
	if got := BgHex("#00ff00", "x"); got != "\x1b[48;2;0;255;0mx\x1b[0m" {
		t.Errorf("BgHex = %q", got)
	}
	if got := Hex("bogus", "x"); got != "x" {
		t.Errorf("Hex bad = %q", got)
	}
}

func TestStripAndWidth(t *testing.T) {
	styled := Bold(Red("hello"))
	if got := Strip(styled); got != "hello" {
		t.Errorf("Strip = %q", got)
	}
	if got := VisibleWidth(styled); got != 5 {
		t.Errorf("VisibleWidth = %d, want 5", got)
	}
	if got := VisibleWidth("plain"); got != 5 {
		t.Errorf("VisibleWidth plain = %d", got)
	}
}

func BenchmarkApply(b *testing.B) {
	s := New().Bold().Red()
	for i := 0; i < b.N; i++ {
		_ = s.Apply("hello")
	}
}
