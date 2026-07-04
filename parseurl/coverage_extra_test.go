package parseurl

import (
	"net/http/httptest"
	"testing"
)

// TestParseFallbackToURL covers the branch where RequestURI is empty and Parse
// falls back to req.URL.
func TestParseFallbackToURL(t *testing.T) {
	req := httptest.NewRequest("GET", "/from/url?a=1", nil)
	req.RequestURI = "" // force the fallback path
	u := Parse(req)
	if u == nil {
		t.Fatal("Parse returned nil")
	}
	if u.Path != "/from/url" {
		t.Fatalf("Path = %q, want /from/url", u.Path)
	}
}

// TestParseAbsoluteURI covers ParseRequestURI succeeding on an absolute-form URI.
func TestParseAbsoluteURI(t *testing.T) {
	req := httptest.NewRequest("GET", "/ignored", nil)
	req.RequestURI = "http://example.com/path?x=1"
	u := Parse(req)
	if u == nil {
		t.Fatal("nil")
	}
	if u.Host != "example.com" || u.Path != "/path" {
		t.Fatalf("host=%q path=%q", u.Host, u.Path)
	}
}

// TestParseRelativeFallback covers the url.Parse fallback when ParseRequestURI
// rejects the target (a relative reference is not a valid request URI).
func TestParseRelativeFallback(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	req.RequestURI = "relative/path?y=2"
	u := Parse(req)
	if u == nil {
		t.Fatal("nil")
	}
	if u.Path != "relative/path" {
		t.Fatalf("Path = %q, want relative/path", u.Path)
	}
	if u.RawQuery != "y=2" {
		t.Fatalf("RawQuery = %q", u.RawQuery)
	}
}

// TestOriginalURLDelegates confirms OriginalURL mirrors Parse.
func TestOriginalURLDelegates(t *testing.T) {
	req := httptest.NewRequest("GET", "/orig?z=3", nil)
	u := OriginalURL(req)
	if u == nil || u.Path != "/orig" {
		t.Fatalf("OriginalURL = %v", u)
	}
}

// TestOriginalURLNil covers the nil-request path via OriginalURL/Parse.
func TestOriginalURLNil(t *testing.T) {
	if u := OriginalURL(nil); u != nil {
		t.Fatalf("OriginalURL(nil) = %v, want nil", u)
	}
}

// TestParseStringError covers ParseString's error return.
func TestParseStringError(t *testing.T) {
	if _, err := ParseString("://bad::url"); err == nil {
		t.Fatal("expected error for malformed URL")
	}
	if u, err := ParseString("/ok?a=1"); err != nil || u.Path != "/ok" {
		t.Fatalf("ParseString ok: u=%v err=%v", u, err)
	}
}
