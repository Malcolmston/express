package cookie

import (
	"strings"
	"testing"
	"time"
)

func TestParseMultiple(t *testing.T) {
	m := Parse("a=1; b=2; c=3")
	if m["a"] != "1" || m["b"] != "2" || m["c"] != "3" {
		t.Fatalf("Parse = %+v", m)
	}
}

func TestParseDecodesValues(t *testing.T) {
	m := Parse("name=John%20Doe; path=%2Ffoo%2Fbar")
	if m["name"] != "John Doe" {
		t.Fatalf("name = %q, want %q", m["name"], "John Doe")
	}
	if m["path"] != "/foo/bar" {
		t.Fatalf("path = %q, want %q", m["path"], "/foo/bar")
	}
}

func TestParseQuotedValue(t *testing.T) {
	m := Parse(`foo="bar"`)
	if m["foo"] != "bar" {
		t.Fatalf("foo = %q, want %q", m["foo"], "bar")
	}
}

func TestParseFirstWins(t *testing.T) {
	m := Parse("a=1; a=2")
	if m["a"] != "1" {
		t.Fatalf("a = %q, want %q", m["a"], "1")
	}
}

func TestParseEmpty(t *testing.T) {
	if m := Parse(""); len(m) != 0 {
		t.Fatalf("Parse(\"\") = %+v, want empty", m)
	}
}

func TestSerializeBasic(t *testing.T) {
	got, err := Serialize("foo", "bar", nil)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if got != "foo=bar" {
		t.Fatalf("Serialize = %q, want %q", got, "foo=bar")
	}
}

func TestSerializeEncodesValue(t *testing.T) {
	got, err := Serialize("foo", "bar baz", nil)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if got != "foo=bar%20baz" {
		t.Fatalf("Serialize = %q, want %q", got, "foo=bar%20baz")
	}
}

func TestSerializeAttributes(t *testing.T) {
	exp := time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)
	got, err := Serialize("sid", "abc", &Options{
		Path:     "/",
		Domain:   "example.com",
		Expires:  exp,
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: true,
		SameSite: "lax",
	})
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	for _, want := range []string{
		"sid=abc",
		"; Path=/",
		"; Domain=example.com",
		"; Expires=Sat, 04 Jul 2026 12:00:00 GMT",
		"; Max-Age=3600",
		"; HttpOnly",
		"; Secure",
		"; SameSite=Lax",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Serialize = %q, missing %q", got, want)
		}
	}
}

func TestSerializeSameSiteVariants(t *testing.T) {
	cases := map[string]string{
		"strict": "; SameSite=Strict",
		"none":   "; SameSite=None",
		"lax":    "; SameSite=Lax",
	}
	for in, want := range cases {
		got, err := Serialize("a", "b", &Options{SameSite: in})
		if err != nil {
			t.Fatalf("Serialize(%q) error: %v", in, err)
		}
		if !strings.Contains(got, want) {
			t.Fatalf("Serialize(%q) = %q, missing %q", in, got, want)
		}
	}
}

func TestSerializeMaxAgeDelete(t *testing.T) {
	got, err := Serialize("a", "b", &Options{MaxAge: -1})
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if !strings.Contains(got, "; Max-Age=0") {
		t.Fatalf("Serialize = %q, want Max-Age=0", got)
	}
}

func TestSerializeInvalidName(t *testing.T) {
	if _, err := Serialize("foo bar", "baz", nil); err == nil {
		t.Fatal("expected error for invalid name")
	}
}

func TestSerializeInvalidSameSite(t *testing.T) {
	if _, err := Serialize("a", "b", &Options{SameSite: "bogus"}); err == nil {
		t.Fatal("expected error for invalid sameSite")
	}
}

func TestRoundTrip(t *testing.T) {
	value := "hello world/@=#?"
	ser, err := Serialize("session", value, nil)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	m := Parse(ser)
	if m["session"] != value {
		t.Fatalf("round-trip = %q, want %q", m["session"], value)
	}
}
