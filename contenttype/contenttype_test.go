package contenttype

import (
	"reflect"
	"testing"
)

func TestParseSimple(t *testing.T) {
	ct, err := Parse("text/html")
	if err != nil {
		t.Fatal(err)
	}
	if ct.Type != "text/html" {
		t.Fatalf("got type %q", ct.Type)
	}
	if len(ct.Parameters) != 0 {
		t.Fatalf("expected no parameters, got %v", ct.Parameters)
	}
}

func TestParseWithParams(t *testing.T) {
	ct, err := Parse("text/html; charset=utf-8")
	if err != nil {
		t.Fatal(err)
	}
	if ct.Type != "text/html" {
		t.Fatalf("got type %q", ct.Type)
	}
	if ct.Parameters["charset"] != "utf-8" {
		t.Fatalf("got charset %q", ct.Parameters["charset"])
	}
}

func TestParseLowercases(t *testing.T) {
	ct, err := Parse("TEXT/HTML; Charset=UTF-8")
	if err != nil {
		t.Fatal(err)
	}
	if ct.Type != "text/html" {
		t.Fatalf("type not lowercased: %q", ct.Type)
	}
	if _, ok := ct.Parameters["charset"]; !ok {
		t.Fatalf("param name not lowercased: %v", ct.Parameters)
	}
	// Value preserves case.
	if ct.Parameters["charset"] != "UTF-8" {
		t.Fatalf("value should preserve case, got %q", ct.Parameters["charset"])
	}
}

func TestParseQuoted(t *testing.T) {
	ct, err := Parse(`text/html; foo="bar baz"`)
	if err != nil {
		t.Fatal(err)
	}
	if ct.Parameters["foo"] != "bar baz" {
		t.Fatalf("got %q", ct.Parameters["foo"])
	}
}

func TestParseQuotedEscape(t *testing.T) {
	ct, err := Parse(`text/html; foo="a\"b\\c"`)
	if err != nil {
		t.Fatal(err)
	}
	if ct.Parameters["foo"] != `a"b\c` {
		t.Fatalf("got %q, want %q", ct.Parameters["foo"], `a"b\c`)
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "text", "text/", "/html", "text/html; =utf-8", "text/html; charset"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) expected error", in)
		}
	}
}

func TestFormatSimple(t *testing.T) {
	got, err := Format(ContentType{Type: "text/html"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "text/html" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatWithParams(t *testing.T) {
	got, err := Format(ContentType{
		Type:       "text/html",
		Parameters: map[string]string{"charset": "utf-8"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "text/html; charset=utf-8" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatQuoting(t *testing.T) {
	got, err := Format(ContentType{
		Type:       "text/html",
		Parameters: map[string]string{"foo": `bar "baz"`},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != `text/html; foo="bar \"baz\""` {
		t.Fatalf("got %q", got)
	}
}

func TestFormatSortedParams(t *testing.T) {
	got, err := Format(ContentType{
		Type:       "text/html",
		Parameters: map[string]string{"z": "1", "a": "2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "text/html; a=2; z=1" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatInvalid(t *testing.T) {
	if _, err := Format(ContentType{Type: "notatype"}); err == nil {
		t.Error("expected error for invalid type")
	}
	if _, err := Format(ContentType{
		Type:       "text/html",
		Parameters: map[string]string{"bad name": "x"},
	}); err == nil {
		t.Error("expected error for invalid param name")
	}
}

func TestRoundTrip(t *testing.T) {
	inputs := []string{
		"text/html",
		"text/html; charset=utf-8",
		`application/json; foo="bar baz"`,
	}
	for _, in := range inputs {
		ct, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		out, err := Format(ct)
		if err != nil {
			t.Fatalf("Format: %v", err)
		}
		ct2, err := Parse(out)
		if err != nil {
			t.Fatalf("re-Parse(%q): %v", out, err)
		}
		if !reflect.DeepEqual(ct, ct2) {
			t.Errorf("round-trip mismatch for %q: %+v vs %+v", in, ct, ct2)
		}
	}
}
