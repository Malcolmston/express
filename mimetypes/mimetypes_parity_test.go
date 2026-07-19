package mimetypes

import "testing"

// Parity tests derived directly from the upstream jshttp/mime-types test suite.
//
// Upstream source (vectors transcribed verbatim from THEIR assertions):
//
//	https://raw.githubusercontent.com/jshttp/mime-types/master/test/test.js
//	https://raw.githubusercontent.com/jshttp/mime-types/master/index.js
//
// The Go port ships a curated subset of mime-db rather than the full generated
// database, so upstream vectors that depend on types absent from that subset
// (e.g. contentType("jade")) are documented as known gaps in
// TestParityKnownGaps rather than asserted as passing. Every other upstream
// vector that touches a type/extension present in the curated table is asserted
// here. Where the Node API returns `false`, the Go port returns ok==false.

// TestParityCharset mirrors upstream describe('.charset(type)').
// Note: upstream returns the charset in canonical casing "UTF-8".
func TestParityCharset(t *testing.T) {
	utf8 := []string{
		"application/json",          // -> 'UTF-8'
		"application/json; foo=bar", // -> 'UTF-8'
		"application/javascript",    // -> 'UTF-8'
		"application/JavaScript",    // -> 'UTF-8' (case-insensitive)
		"text/html",                 // -> 'UTF-8'
		"TEXT/HTML",                 // -> 'UTF-8' (case-insensitive)
		"text/x-bogus",              // -> 'UTF-8' (any text/*)
	}
	for _, in := range utf8 {
		got, ok := Charset(in)
		if !ok || got != "UTF-8" {
			t.Errorf("Charset(%q) = %q,%v; want %q,true", in, got, ok, "UTF-8")
		}
	}

	// upstream: false
	falses := []string{
		"application/x-bogus",
		"application/octet-stream",
		"", // invalid argument
	}
	for _, in := range falses {
		if got, ok := Charset(in); ok {
			t.Errorf("Charset(%q) = %q,%v; want \"\",false", in, got, ok)
		}
	}
}

// TestParityContentTypeExtension mirrors upstream describe('.contentType(extension)').
func TestParityContentTypeExtension(t *testing.T) {
	cases := map[string]string{
		"html":  "text/html; charset=utf-8",
		".html": "text/html; charset=utf-8",
		"json":  "application/json; charset=utf-8",
	}
	for in, want := range cases {
		got, ok := ContentType(in)
		if !ok || got != want {
			t.Errorf("ContentType(%q) = %q,%v; want %q,true", in, got, ok, want)
		}
	}

	// upstream: false for unknown extensions / empty
	for _, in := range []string{"bogus", ""} {
		if got, ok := ContentType(in); ok {
			t.Errorf("ContentType(%q) = %q,%v; want \"\",false", in, got, ok)
		}
	}
}

// TestParityContentTypeType mirrors upstream describe('.contentType(type)').
func TestParityContentTypeType(t *testing.T) {
	cases := map[string]string{
		"application/json":              "application/json; charset=utf-8",
		"application/json; foo=bar":     "application/json; foo=bar; charset=utf-8",
		"TEXT/HTML":                     "TEXT/HTML; charset=utf-8",
		"text/html":                     "text/html; charset=utf-8",
		"text/html; charset=iso-8859-1": "text/html; charset=iso-8859-1", // unaltered
		"application/x-bogus":           "application/x-bogus",           // unknown type returned as-is
	}
	for in, want := range cases {
		got, ok := ContentType(in)
		if !ok || got != want {
			t.Errorf("ContentType(%q) = %q,%v; want %q,true", in, got, ok, want)
		}
	}
}

// TestParityExtension mirrors upstream describe('.extension(type)').
func TestParityExtension(t *testing.T) {
	html := []string{
		"text/html",
		" text/html",                // leading space
		"text/html ",                // trailing space
		"text/html;charset=UTF-8",   // params, no space
		"text/HTML; charset=UTF-8",  // case + params
		"text/html; charset=UTF-8",  // params
		"text/html; charset=UTF-8 ", // params + trailing space
		"text/html ; charset=UTF-8", // space before semicolon
	}
	for _, in := range html {
		got, ok := Extension(in)
		if !ok || got != "html" {
			t.Errorf("Extension(%q) = %q,%v; want %q,true", in, got, ok, "html")
		}
	}

	// upstream: false
	for _, in := range []string{"application/x-bogus", "bogus", ""} {
		if got, ok := Extension(in); ok {
			t.Errorf("Extension(%q) = %q,%v; want \"\",false", in, got, ok)
		}
	}
}

// TestParityLookupExtension mirrors upstream describe('.lookup(extension)').
func TestParityLookupExtension(t *testing.T) {
	cases := map[string]string{
		".html": "text/html",
		".js":   "text/javascript", // current mime-db via mime-score
		".json": "application/json",
		".rtf":  "application/rtf",
		".txt":  "text/plain",
		".xml":  "application/xml",
		".mp4":  "video/mp4",
		"html":  "text/html", // without leading dot
		"xml":   "application/xml",
		"HTML":  "text/html", // case-insensitive
		".Xml":  "application/xml",
	}
	for in, want := range cases {
		got, ok := Lookup(in)
		if !ok || got != want {
			t.Errorf("Lookup(%q) = %q,%v; want %q,true", in, got, ok, want)
		}
	}

	// upstream: false
	for _, in := range []string{".bogus", "bogus", ""} {
		if got, ok := Lookup(in); ok {
			t.Errorf("Lookup(%q) = %q,%v; want \"\",false", in, got, ok)
		}
	}
}

// TestParityLookupPath mirrors upstream describe('.lookup(path)'), including the
// 'path with dotfile' sub-suite.
func TestParityLookupPath(t *testing.T) {
	cases := map[string]string{
		"page.html":               "text/html",
		"path/to/page.html":       "text/html",
		"path\\to\\page.html":     "text/html",
		"/path/to/page.html":      "text/html",
		"C:\\path\\to\\page.html": "text/html",
		"/path/to/PAGE.HTML":      "text/html", // case-insensitive
		"C:\\path\\to\\PAGE.HTML": "text/html",
		"/path/to/.config.json":   "application/json", // dotfile WITH extension
		".config.json":            "application/json", // dotfile with extension, no path
	}
	for in, want := range cases {
		got, ok := Lookup(in)
		if !ok || got != want {
			t.Errorf("Lookup(%q) = %q,%v; want %q,true", in, got, ok, want)
		}
	}

	// upstream: false
	falses := []string{
		"/path/to/file.bogus", // unknown extension
		"/path/to/json",       // no extension
		"/path/to/.json",      // extension-less dotfile
	}
	for _, in := range falses {
		if got, ok := Lookup(in); ok {
			t.Errorf("Lookup(%q) = %q,%v; want \"\",false", in, got, ok)
		}
	}
}

// TestParityKnownGaps documents upstream vectors that the curated Go port does
// not cover because the underlying type is absent from its bundled subset of
// mime-db (the port intentionally ships a curated table, not the full generated
// database). These assert the port's ACTUAL current behavior and record, in
// comments, what upstream returns, so the gap is visible without failing CI.
func TestParityKnownGaps(t *testing.T) {
	// upstream: contentType('jade') -> 'text/jade; charset=utf-8'
	// The "jade" extension / "text/jade" type are not in the curated table.
	if got, ok := ContentType("jade"); ok {
		t.Errorf("ContentType(%q) = %q,%v; curated port is expected to lack \"jade\" (upstream: text/jade; charset=utf-8)", "jade", got, ok)
	}
}
