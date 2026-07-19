package sqlstring

import (
	"testing"
	"time"
)

func TestEscapeScalars(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, "NULL"},
		{true, "true"},
		{false, "false"},
		{42, "42"},
		{int64(-7), "-7"},
		{uint(9), "9"},
		{3.5, "3.5"},
		{"hello", "'hello'"},
	}
	for _, c := range cases {
		if got := Escape(c.in); got != c.want {
			t.Errorf("Escape(%v) = %q want %q", c.in, got, c.want)
		}
	}
}

func TestEscapeStringSpecials(t *testing.T) {
	got := Escape("a'b\"c\\d\ne\tf\x00g\x1ah\r\bi")
	want := `'a\'b\"c\\d\ne\tf\0g\Zh\r\bi'`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestEscapeBytes(t *testing.T) {
	if got := Escape([]byte{0xde, 0xad, 0xbe, 0xef}); got != "X'deadbeef'" {
		t.Errorf("got %q", got)
	}
}

func TestEscapeTime(t *testing.T) {
	tm := time.Date(2026, 7, 4, 13, 5, 9, 0, time.UTC)
	if got := Escape(tm); got != "'2026-07-04 13:05:09.000'" {
		t.Errorf("got %q", got)
	}
}

func TestEscapeSlice(t *testing.T) {
	// Top-level lists are not parenthesized (upstream arrayToList).
	if got := Escape([]any{1, "two", nil}); got != "1, 'two', NULL" {
		t.Errorf("got %q", got)
	}
	if got := Escape([]int{1, 2, 3}); got != "1, 2, 3" {
		t.Errorf("got %q", got)
	}
	// Nested lists are grouped in parentheses.
	if got := Escape([]any{[]int{1, 2}, []int{3, 4}}); got != "(1, 2), (3, 4)" {
		t.Errorf("got %q", got)
	}
}

func TestEscapePointer(t *testing.T) {
	x := 5
	if got := Escape(&x); got != "5" {
		t.Errorf("got %q", got)
	}
	var p *int
	if got := Escape(p); got != "NULL" {
		t.Errorf("nil ptr got %q", got)
	}
}

func TestEscapeID(t *testing.T) {
	if got := EscapeID("foo"); got != "`foo`" {
		t.Errorf("got %q", got)
	}
	if got := EscapeID("has`tick"); got != "`has``tick`" {
		t.Errorf("got %q", got)
	}
	if got := EscapeID("table.col"); got != "`table`.`col`" {
		t.Errorf("got %q", got)
	}
}

func TestFormat(t *testing.T) {
	got := Format("SELECT * FROM ?? WHERE id = ? AND name = ?", []any{"users", 5, "bob"})
	want := "SELECT * FROM `users` WHERE id = 5 AND name = 'bob'"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestFormatLeftovers(t *testing.T) {
	// Fewer args than placeholders leaves extras in place.
	if got := Format("a=? b=?", []any{1}); got != "a=1 b=?" {
		t.Errorf("got %q", got)
	}
	// No args returns sql unchanged.
	if got := Format("no placeholders", nil); got != "no placeholders" {
		t.Errorf("got %q", got)
	}
}

func TestFormatIDNonString(t *testing.T) {
	if got := Format("x ??", []any{123}); got != "x `123`" {
		t.Errorf("got %q", got)
	}
}
