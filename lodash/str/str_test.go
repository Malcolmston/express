package str

import (
	"reflect"
	"regexp"
	"testing"
)

func TestUpperLowerFirst(t *testing.T) {
	if got := UpperFirst("fred"); got != "Fred" {
		t.Errorf("UpperFirst(fred) = %q", got)
	}
	if got := UpperFirst("FRED"); got != "FRED" {
		t.Errorf("UpperFirst(FRED) = %q", got)
	}
	if got := LowerFirst("Fred"); got != "fred" {
		t.Errorf("LowerFirst(Fred) = %q", got)
	}
	if got := LowerFirst("FRED"); got != "fRED" {
		t.Errorf("LowerFirst(FRED) = %q", got)
	}
	if got := UpperFirst(""); got != "" {
		t.Errorf("UpperFirst empty = %q", got)
	}
}

func TestCapitalize(t *testing.T) {
	cases := map[string]string{
		"FRED":  "Fred",
		"fred":  "Fred",
		"fReD":  "Fred",
		"hello": "Hello",
	}
	for in, want := range cases {
		if got := Capitalize(in); got != want {
			t.Errorf("Capitalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWords(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"fooBar", []string{"foo", "Bar"}},
		{"foo_bar-baz", []string{"foo", "bar", "baz"}},
		{"XMLHttpRequest", []string{"XML", "Http", "Request"}},
		{"fred, barney, & pebbles", []string{"fred", "barney", "pebbles"}},
		{"--foo-bar--", []string{"foo", "bar"}},
		{"", nil},
	}
	for _, c := range cases {
		if got := Words(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Words(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestCaseConverters(t *testing.T) {
	if got := CamelCase("Foo Bar"); got != "fooBar" {
		t.Errorf("CamelCase = %q", got)
	}
	if got := CamelCase("__FOO_BAR__"); got != "fooBar" {
		t.Errorf("CamelCase = %q", got)
	}
	if got := KebabCase("fooBar"); got != "foo-bar" {
		t.Errorf("KebabCase = %q", got)
	}
	if got := KebabCase("__FOO_BAR__"); got != "foo-bar" {
		t.Errorf("KebabCase = %q", got)
	}
	if got := SnakeCase("Foo Bar"); got != "foo_bar" {
		t.Errorf("SnakeCase = %q", got)
	}
	if got := StartCase("--foo-bar--"); got != "Foo Bar" {
		t.Errorf("StartCase = %q", got)
	}
	if got := StartCase("fooBar"); got != "Foo Bar" {
		t.Errorf("StartCase = %q", got)
	}
	if got := StartCase("XMLHttp"); got != "XML Http" {
		t.Errorf("StartCase = %q", got)
	}
	if got := LowerCase("--Foo-Bar--"); got != "foo bar" {
		t.Errorf("LowerCase = %q", got)
	}
	if got := UpperCase("--foo-bar--"); got != "FOO BAR" {
		t.Errorf("UpperCase = %q", got)
	}
}

func TestPad(t *testing.T) {
	if got := Pad("abc", 8, "_-"); got != "_-abc_-_" {
		t.Errorf("Pad = %q", got)
	}
	if got := Pad("abc", 3, ""); got != "abc" {
		t.Errorf("Pad noop = %q", got)
	}
	if got := PadStart("abc", 6, ""); got != "   abc" {
		t.Errorf("PadStart = %q", got)
	}
	if got := PadStart("abc", 6, "_-"); got != "_-_abc" {
		t.Errorf("PadStart = %q", got)
	}
	if got := PadEnd("abc", 6, ""); got != "abc   " {
		t.Errorf("PadEnd = %q", got)
	}
	if got := PadEnd("abc", 6, "_-"); got != "abc_-_" {
		t.Errorf("PadEnd = %q", got)
	}
}

func TestRepeat(t *testing.T) {
	if got := Repeat("*", 3); got != "***" {
		t.Errorf("Repeat = %q", got)
	}
	if got := Repeat("abc", 2); got != "abcabc" {
		t.Errorf("Repeat = %q", got)
	}
	if got := Repeat("abc", 0); got != "" {
		t.Errorf("Repeat 0 = %q", got)
	}
}

func TestTrim(t *testing.T) {
	if got := Trim("  abc  ", ""); got != "abc" {
		t.Errorf("Trim = %q", got)
	}
	if got := Trim("-_-abc-_-", "_-"); got != "abc" {
		t.Errorf("Trim chars = %q", got)
	}
	if got := TrimStart("-_-abc-_-", "_-"); got != "abc-_-" {
		t.Errorf("TrimStart = %q", got)
	}
	if got := TrimEnd("-_-abc-_-", "_-"); got != "-_-abc" {
		t.Errorf("TrimEnd = %q", got)
	}
	if got := TrimStart("  abc  ", ""); got != "abc  " {
		t.Errorf("TrimStart ws = %q", got)
	}
	if got := TrimEnd("  abc  ", ""); got != "  abc" {
		t.Errorf("TrimEnd ws = %q", got)
	}
}

func TestStartsEndsWith(t *testing.T) {
	if !StartsWith("abc", "a", 0) {
		t.Error("StartsWith a@0")
	}
	if !StartsWith("abc", "b", 1) {
		t.Error("StartsWith b@1")
	}
	if StartsWith("abc", "b", 0) {
		t.Error("StartsWith b@0 should be false")
	}
	if !EndsWith("abc", "c", -1) {
		t.Error("EndsWith c")
	}
	if !EndsWith("abc", "b", 2) {
		t.Error("EndsWith b@2")
	}
	if EndsWith("abc", "b", -1) {
		t.Error("EndsWith b should be false")
	}
}

func TestEscapeUnescape(t *testing.T) {
	in := "fred, barney, & pebbles"
	esc := "fred, barney, &amp; pebbles"
	if got := Escape(in); got != esc {
		t.Errorf("Escape = %q", got)
	}
	if got := Unescape(esc); got != in {
		t.Errorf("Unescape = %q", got)
	}
	full := `<a href="x">'y'</a>`
	if got := Unescape(Escape(full)); got != full {
		t.Errorf("round trip = %q", got)
	}
	if got := Escape("<>"); got != "&lt;&gt;" {
		t.Errorf("Escape <> = %q", got)
	}
}

func TestTruncate(t *testing.T) {
	s := "hi-diddly-ho there, neighborino"
	if got := Truncate(s, TruncateOptions{}); got != "hi-diddly-ho there, neighbo..." {
		t.Errorf("Truncate default = %q", got)
	}
	if got := Truncate(s, TruncateOptions{Length: 24, Separator: " "}); got != "hi-diddly-ho there,..." {
		t.Errorf("Truncate sep = %q", got)
	}
	if got := Truncate(s, TruncateOptions{Length: 24, Omission: " [...]"}); got != "hi-diddly-ho there [...]" {
		t.Errorf("Truncate omission = %q", got)
	}
	if got := Truncate("short", TruncateOptions{}); got != "short" {
		t.Errorf("Truncate noop = %q", got)
	}
	re := regexp.MustCompile(`,? +`)
	if got := Truncate(s, TruncateOptions{Length: 24, SeparatorRegexp: re}); got != "hi-diddly-ho there..." {
		t.Errorf("Truncate regexp = %q", got)
	}
}

func TestReplace(t *testing.T) {
	if got := Replace("Hi Fred", "Fred", "Barney"); got != "Hi Barney" {
		t.Errorf("Replace = %q", got)
	}
	if got := Replace("aaa", "a", "b"); got != "baa" {
		t.Errorf("Replace first only = %q", got)
	}
}

func TestParseInt(t *testing.T) {
	cases := []struct {
		s     string
		radix int
		want  int
	}{
		{"08", 10, 8},
		{"10", 2, 2},
		{"0x1A", 0, 26},
		{"0x1A", 16, 26},
		{"42px", 10, 42},
		{"  -15  ", 10, -15},
		{"ff", 16, 255},
		{"xyz", 10, 0},
	}
	for _, c := range cases {
		if got := ParseInt(c.s, c.radix); got != c.want {
			t.Errorf("ParseInt(%q,%d) = %d, want %d", c.s, c.radix, got, c.want)
		}
	}
}

func TestDeburr(t *testing.T) {
	if got := Deburr("déjà vu"); got != "deja vu" {
		t.Errorf("Deburr = %q", got)
	}
	if got := Deburr("Æther"); got != "Aether" {
		t.Errorf("Deburr Ae = %q", got)
	}
	// combining marks are dropped
	if got := Deburr("é"); got != "e" {
		t.Errorf("Deburr combining = %q", got)
	}
	// deburr feeds the case converters
	if got := CamelCase("déjà vu"); got != "dejaVu" {
		t.Errorf("CamelCase deburr = %q", got)
	}
}
