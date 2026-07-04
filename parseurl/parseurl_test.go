package parseurl

import (
	"net/http/httptest"
	"testing"
)

func TestParsePathAndQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo/bar?baz=1&qux=2", nil)
	u := Parse(req)
	if u == nil {
		t.Fatal("Parse returned nil")
	}
	if u.Path != "/foo/bar" {
		t.Fatalf("Path = %q, want %q", u.Path, "/foo/bar")
	}
	if u.RawQuery != "baz=1&qux=2" {
		t.Fatalf("RawQuery = %q, want %q", u.RawQuery, "baz=1&qux=2")
	}
	if got := u.Query().Get("baz"); got != "1" {
		t.Fatalf("Query baz = %q, want %q", got, "1")
	}
}

func TestParseNoQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/plain", nil)
	u := Parse(req)
	if u.Path != "/plain" {
		t.Fatalf("Path = %q, want %q", u.Path, "/plain")
	}
	if u.RawQuery != "" {
		t.Fatalf("RawQuery = %q, want empty", u.RawQuery)
	}
}

func TestParseEncodedPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo%20bar?a=b%20c", nil)
	u := Parse(req)
	if u.Path != "/foo bar" {
		t.Fatalf("Path = %q, want %q", u.Path, "/foo bar")
	}
	if u.EscapedPath() != "/foo%20bar" {
		t.Fatalf("EscapedPath = %q, want %q", u.EscapedPath(), "/foo%20bar")
	}
}

func TestParseNil(t *testing.T) {
	if Parse(nil) != nil {
		t.Fatal("Parse(nil) should return nil")
	}
}

func TestOriginalURL(t *testing.T) {
	req := httptest.NewRequest("GET", "/orig?x=1", nil)
	u := OriginalURL(req)
	if u == nil || u.Path != "/orig" || u.RawQuery != "x=1" {
		t.Fatalf("OriginalURL = %+v", u)
	}
}

func TestParseString(t *testing.T) {
	u, err := ParseString("http://example.com/a/b?c=d")
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}
	if u.Host != "example.com" || u.Path != "/a/b" || u.RawQuery != "c=d" {
		t.Fatalf("ParseString = %+v", u)
	}
}
