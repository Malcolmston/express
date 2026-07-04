package wordwrap

import (
	"strings"
	"testing"
)

func TestWrapDefaults(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog and then runs away quickly"
	got := Wrap(text, NewOptions())
	lines := strings.Split(got, "\n")
	for _, ln := range lines {
		if !strings.HasPrefix(ln, DefaultIndent) {
			t.Errorf("line %q missing indent", ln)
		}
		content := strings.TrimPrefix(ln, DefaultIndent)
		if len(content) > DefaultWidth {
			t.Errorf("line content %q exceeds width %d", content, DefaultWidth)
		}
	}
	if len(lines) < 2 {
		t.Errorf("expected multiple lines, got %d", len(lines))
	}
}

func TestWrapReconstruct(t *testing.T) {
	text := "one two three four five six seven eight nine ten"
	opts := Options{Width: 10}
	got := Wrap(text, opts)
	// No indent, so joining lines back with spaces should recover the words.
	rejoined := strings.Join(strings.Fields(got), " ")
	if rejoined != text {
		t.Errorf("reconstruct = %q, want %q", rejoined, text)
	}
}

func TestWrapWidth(t *testing.T) {
	text := "aaa bbb ccc ddd"
	got := Wrap(text, Options{Width: 7})
	want := "aaa bbb\nccc ddd"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapZeroWidthDefaults(t *testing.T) {
	// Width 0 should behave as DefaultWidth.
	text := "short text"
	got := Wrap(text, Options{})
	if got != "short text" {
		t.Errorf("got %q, want %q", got, "short text")
	}
}

func TestWrapCut(t *testing.T) {
	text := "supercalifragilistic"
	got := Wrap(text, Options{Width: 5, Cut: true})
	want := "super\ncalif\nragil\nistic"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapNoCutLongWord(t *testing.T) {
	text := "supercalifragilistic word"
	got := Wrap(text, Options{Width: 5})
	want := "supercalifragilistic\nword"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapPreserveNewlines(t *testing.T) {
	text := "line one\nline two"
	got := Wrap(text, Options{Width: 50})
	want := "line one\nline two"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapCustomNewlineAndIndent(t *testing.T) {
	text := "aaa bbb ccc"
	got := Wrap(text, Options{Width: 3, Indent: ">>", Newline: "|"})
	want := ">>aaa|>>bbb|>>ccc"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapTrimTrailing(t *testing.T) {
	// A blank paragraph produces an indented blank line; trimming removes the
	// trailing indent whitespace.
	text := "aaa\n\nbbb"
	got := Wrap(text, Options{Width: 10, Indent: "  ", TrimTrailing: true})
	want := "  aaa\n\n  bbb"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapBlankLineNoTrim(t *testing.T) {
	text := "aaa\n\nbbb"
	got := Wrap(text, Options{Width: 10, Indent: "  "})
	want := "  aaa\n  \n  bbb"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNewOptions(t *testing.T) {
	o := NewOptions()
	if o.Width != 50 || o.Indent != "  " || o.Newline != "\n" {
		t.Errorf("unexpected defaults: %+v", o)
	}
}
