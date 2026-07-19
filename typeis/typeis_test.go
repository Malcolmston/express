package typeis

import "testing"

func TestIsExact(t *testing.T) {
	got, ok := Is("application/json", "application/json")
	if !ok || got != "application/json" {
		t.Fatalf("Is = %q,%v", got, ok)
	}
}

func TestIsWithParams(t *testing.T) {
	// Upstream returns the candidate as supplied ("json") for a non-wildcard
	// match; parameters on the value are stripped before matching.
	got, ok := Is("application/json; charset=utf-8", "json")
	if !ok || got != "json" {
		t.Fatalf("Is = %q,%v", got, ok)
	}
}

func TestIsShorthand(t *testing.T) {
	cases := []struct {
		ct        string
		shorthand string
		want      string
	}{
		// A non-wildcard candidate is echoed back exactly as supplied
		// (upstream type-is convention).
		{"application/json", "json", "json"},
		{"text/html", "html", "html"},
		{"application/x-www-form-urlencoded", "urlencoded", "urlencoded"},
		{"multipart/form-data", "multipart", "multipart"},
	}
	for _, c := range cases {
		got, ok := Is(c.ct, c.shorthand)
		if !ok || got != c.want {
			t.Errorf("Is(%q,%q) = %q,%v; want %q", c.ct, c.shorthand, got, ok, c.want)
		}
	}
}

func TestIsWildcard(t *testing.T) {
	if _, ok := Is("image/png", "*/*"); !ok {
		t.Error("*/* should match anything")
	}
	if _, ok := Is("text/html", "text/*"); !ok {
		t.Error("text/* should match text/html")
	}
	if _, ok := Is("application/json", "text/*"); ok {
		t.Error("text/* should not match application/json")
	}
	if _, ok := Is("application/json", "*/json"); !ok {
		t.Error("*/json should match application/json")
	}
}

func TestIsSuffix(t *testing.T) {
	got, ok := Is("application/vnd.api+json", "+json")
	if !ok {
		t.Fatal("+json should match application/vnd.api+json")
	}
	_ = got
	if _, ok := Is("application/xml", "+json"); ok {
		t.Error("+json should not match application/xml")
	}
	if _, ok := Is("application/vnd.api+json", "json"); ok {
		t.Error("json shorthand should not match a +json suffix type")
	}
}

func TestIsNoMatch(t *testing.T) {
	if _, ok := Is("text/plain", "json", "html"); ok {
		t.Error("text/plain should not match json or html")
	}
}

func TestIsNoOffers(t *testing.T) {
	got, ok := Is("text/html; charset=utf-8")
	if !ok || got != "text/html" {
		t.Errorf("Is with no offers = %q,%v", got, ok)
	}
}

func TestIsEmptyContentType(t *testing.T) {
	if _, ok := Is("", "json"); ok {
		t.Error("empty content-type should not match")
	}
}

func TestMatch(t *testing.T) {
	cases := []struct {
		expected, actual string
		want             bool
	}{
		{"application/json", "application/json", true},
		{"text/*", "text/html", true},
		{"text/*", "application/json", false},
		{"*/*", "anything/here", true},
		{"*/json", "application/json", true},
		{"*/*+json", "application/vnd.api+json", true},
		{"application/json", "application/xml", false},
	}
	for _, c := range cases {
		if got := Match(c.expected, c.actual); got != c.want {
			t.Errorf("Match(%q,%q) = %v; want %v", c.expected, c.actual, got, c.want)
		}
	}
}
