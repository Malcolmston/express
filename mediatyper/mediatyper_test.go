package mediatyper

import "testing"

func TestParseSimple(t *testing.T) {
	m, err := Parse("text/html")
	if err != nil {
		t.Fatal(err)
	}
	if m.Type != "text" || m.Subtype != "html" || m.Suffix != "" {
		t.Fatalf("got %+v", m)
	}
}

func TestParseSuffix(t *testing.T) {
	m, err := Parse("application/vnd.api+json")
	if err != nil {
		t.Fatal(err)
	}
	if m.Type != "application" || m.Subtype != "vnd.api" || m.Suffix != "json" {
		t.Fatalf("got %+v", m)
	}
}

func TestParseParams(t *testing.T) {
	m, err := Parse("application/vnd.api+json; charset=utf-8")
	if err != nil {
		t.Fatal(err)
	}
	if m.Parameters["charset"] != "utf-8" {
		t.Fatalf("charset = %q", m.Parameters["charset"])
	}
}

func TestParseQuotedParam(t *testing.T) {
	m, err := Parse(`text/plain; foo="a; b"`)
	if err != nil {
		t.Fatal(err)
	}
	if m.Parameters["foo"] != "a; b" {
		t.Fatalf("foo = %q", m.Parameters["foo"])
	}
}

func TestParseInvalid(t *testing.T) {
	bad := []string{"", "noslash", "text/", "/html", "text/ht ml", "te xt/html"}
	for _, s := range bad {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) expected error", s)
		}
	}
}

func TestFormatSimple(t *testing.T) {
	got, err := Format(MediaType{Type: "text", Subtype: "html"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "text/html" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatSuffixAndParams(t *testing.T) {
	got, err := Format(MediaType{
		Type:       "application",
		Subtype:    "vnd.api",
		Suffix:     "json",
		Parameters: map[string]string{"charset": "utf-8"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "application/vnd.api+json; charset=utf-8" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatInvalid(t *testing.T) {
	if _, err := Format(MediaType{Type: "te xt", Subtype: "html"}); err == nil {
		t.Error("expected error for invalid type")
	}
	if _, err := Format(MediaType{Type: "text", Subtype: ""}); err == nil {
		t.Error("expected error for empty subtype")
	}
}

func TestRoundTrip(t *testing.T) {
	inputs := []string{
		"text/html",
		"application/json",
		"application/vnd.api+json; charset=utf-8",
		"image/svg+xml",
	}
	for _, in := range inputs {
		m, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		out, err := Format(m)
		if err != nil {
			t.Fatalf("Format: %v", err)
		}
		if out != in {
			t.Errorf("round-trip %q -> %q", in, out)
		}
	}
}
