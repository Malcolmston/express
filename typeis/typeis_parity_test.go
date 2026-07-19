package typeis

// Upstream-parity tests for the npm "jshttp/type-is" package.
//
// Every vector below is transcribed from the original library's own test
// suite and its return-value contract in index.js:
//
//   https://raw.githubusercontent.com/jshttp/type-is/master/test/test.js
//   https://raw.githubusercontent.com/jshttp/type-is/master/index.js
//
// Mapping of upstream API to this Go port:
//   typeis.is(mediaType, types)   -> Is(mediaType, types...) (string, bool)
//   typeis.match(expected, actual)-> Match(expected, actual) bool
//   typeis.normalize(type)        -> Normalize(type) (string, bool)
//
// Upstream returns a string-or-false union; this port returns (string, bool),
// so an upstream `false`/`null` result maps to ok == false and an upstream
// string result maps to (that string, true). The request-object form of typeis
// and typeis.hasBody depend on a live request and are intentionally omitted by
// this port (see package docs); they are therefore not exercised here.

import "testing"

// TestParityIs covers the describe('typeis.is(mediaType, types)') block, whose
// vectors are identical in behaviour to the request form. Upstream returns the
// candidate as supplied for non-wildcard matches, and the concrete value for
// candidates containing "*" or beginning with "+".
func TestParityIs(t *testing.T) {
	cases := []struct {
		value   string
		types   []string
		want    string
		wantOk  bool
		comment string
	}{
		// should ignore params / casing / invalid types
		{"text/html; charset=utf-8", []string{"text/*"}, "text/html", true, "ignore params"},
		{"text/html ; charset=utf-8", []string{"text/*"}, "text/html", true, "ignore params LWS"},
		{"text/HTML", []string{"text/*"}, "text/html", true, "ignore casing"},
		{"text/html**", []string{"text/*"}, "", false, "fail invalid type"},
		{"text/html", []string{"text/html/"}, "", false, "invalid candidate type"},

		// when no media type is given -> false
		{"", nil, "", false, "no media type"},
		{"", []string{"application/json"}, "", false, "empty value"},

		// given no types -> the mime type
		{"image/png", nil, "image/png", true, "no types returns value"},

		// given one type -> the type or false
		{"image/png", []string{"png"}, "png", true, "extension png"},
		{"image/png", []string{".png"}, ".png", true, "dotted extension"},
		{"image/png", []string{"image/png"}, "image/png", true, "full type"},
		{"image/png", []string{"image/*"}, "image/png", true, "subtype wildcard -> value"},
		{"image/png", []string{"*/png"}, "image/png", true, "type wildcard -> value"},
		{"image/png", []string{"jpeg"}, "", false, "wrong extension"},
		{"image/png", []string{".jpeg"}, "", false, "wrong dotted extension"},
		{"image/png", []string{"image/jpeg"}, "", false, "wrong full type"},
		{"image/png", []string{"text/*"}, "", false, "wrong type wildcard"},
		{"image/png", []string{"*/jpeg"}, "", false, "wrong subtype wildcard"},
		{"image/png", []string{"bogus"}, "", false, "unknown extension"},
		{"image/png", []string{"something/bogus*"}, "", false, "bogus full type"},

		// given multiple types -> first match or false
		{"image/png", []string{"png"}, "png", true, "multi single ext"},
		{"image/png", []string{".png"}, ".png", true, "multi dotted ext"},
		{"image/png", []string{"text/*", "image/*"}, "image/png", true, "multi wildcard second"},
		{"image/png", []string{"image/*", "text/*"}, "image/png", true, "multi wildcard first"},
		{"image/png", []string{"image/*", "image/png"}, "image/png", true, "multi wildcard then exact"},
		{"image/png", []string{"image/png", "image/*"}, "image/png", true, "multi exact then wildcard"},
		{"image/png", []string{"jpeg"}, "", false, "multi no match ext"},
		{"image/png", []string{".jpeg"}, "", false, "multi no match dotted"},
		{"image/png", []string{"text/*", "application/*"}, "", false, "multi no match wildcards"},
		{"image/png", []string{"text/html", "text/plain", "application/json"}, "", false, "multi no match full"},

		// given +suffix
		{"application/vnd+json", []string{"+json"}, "application/vnd+json", true, "suffix -> value"},
		{"application/vnd+json", []string{"application/vnd+json"}, "application/vnd+json", true, "exact suffix type"},
		{"application/vnd+json", []string{"application/*+json"}, "application/vnd+json", true, "subtype wildcard suffix"},
		{"application/vnd+json", []string{"*/vnd+json"}, "application/vnd+json", true, "type wildcard suffix"},
		{"application/vnd+json", []string{"application/json"}, "", false, "suffix vs plain json"},
		{"application/vnd+json", []string{"text/*+json"}, "", false, "suffix wrong type"},

		// given "*/*"
		{"text/html", []string{"*/*"}, "text/html", true, "full wildcard html"},
		{"text/xml", []string{"*/*"}, "text/xml", true, "full wildcard xml"},
		{"application/json", []string{"*/*"}, "application/json", true, "full wildcard json"},
		{"application/vnd+json", []string{"*/*"}, "application/vnd+json", true, "full wildcard suffix"},
		{"bogus", []string{"*/*"}, "", false, "full wildcard invalid value"},

		// urlencoded special
		{"application/x-www-form-urlencoded", []string{"urlencoded"}, "urlencoded", true, "urlencoded"},
		{"application/x-www-form-urlencoded", []string{"json", "urlencoded"}, "urlencoded", true, "urlencoded second"},
		{"application/x-www-form-urlencoded", []string{"urlencoded", "json"}, "urlencoded", true, "urlencoded first"},

		// multipart special
		{"multipart/form-data", []string{"multipart/*"}, "multipart/form-data", true, "multipart wildcard -> value"},
		{"multipart/form-data", []string{"multipart"}, "multipart", true, "multipart special"},
	}
	for _, c := range cases {
		got, ok := Is(c.value, c.types...)
		if ok != c.wantOk || (ok && got != c.want) {
			t.Errorf("Is(%q, %v) = %q,%v; want %q,%v (%s)",
				c.value, c.types, got, ok, c.want, c.wantOk, c.comment)
		}
	}
}

// TestParityIsVariadic covers the flattened-argument call sites from the
// upstream "given multiple types" blocks, e.g. typeis.is('image/png',
// 'image/png', 'image/*') and typeis.is('image/png', '.png').
func TestParityIsVariadic(t *testing.T) {
	if got, ok := Is("image/png", ".png"); !ok || got != ".png" {
		t.Errorf("Is variadic dotted = %q,%v; want .png,true", got, ok)
	}
	if got, ok := Is("image/png", "image/png", "image/*"); !ok || got != "image/png" {
		t.Errorf("Is variadic = %q,%v; want image/png,true", got, ok)
	}
}

// TestParityMatch covers the describe('typeis.match(expected, actual)') block.
func TestParityMatch(t *testing.T) {
	cases := []struct {
		expected, actual string
		want             bool
	}{
		// expected is false
		{"", "text/html", false},
		// exact matching
		{"text/html", "text/html", true},
		{"text/html", "text/plain", false},
		{"text/html", "text/xml", false},
		{"text/html", "application/html", false},
		{"text/html", "text/html+xml", false},
		// type wildcard matching
		{"*/html", "text/html", true},
		{"*/html", "application/html", true},
		{"*/html", "text/xml", false},
		{"*/html", "text/html+xml", false},
		// subtype wildcard matching
		{"text/*", "text/html", true},
		{"text/*", "text/xml", true},
		{"text/*", "text/html+xml", true},
		{"text/*", "application/xml", false},
		// full wildcard matching
		{"*/*", "text/html", true},
		{"*/*", "text/html+xml", true},
		{"*/*+xml", "text/html+xml", true},
		// full wildcard with specific suffix
		{"*/*+xml", "text/html", false},
	}
	for _, c := range cases {
		if got := Match(c.expected, c.actual); got != c.want {
			t.Errorf("Match(%q,%q) = %v; want %v", c.expected, c.actual, got, c.want)
		}
	}
}

// TestParityNormalize covers the describe('typeis.normalize(type)') block.
// Upstream returns false for unmapped extensions; here that is ok == false.
func TestParityNormalize(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantOk bool
	}{
		{"json", "application/json", true},             // media type for extension
		{"+json", "*/*+json", true},                    // expanded wildcard for suffix
		{"application/json", "application/json", true}, // pass through media type
		{"*/*", "*/*", true},                           // pass through wildcard
		{"image/*", "image/*", true},                   // pass through wildcard
		{"unknown", "", false},                         // unmapped extension -> false
		{"urlencoded", "application/x-www-form-urlencoded", true},
		{"multipart", "multipart/*", true},
	}
	for _, c := range cases {
		got, ok := Normalize(c.in)
		if ok != c.wantOk || (ok && got != c.want) {
			t.Errorf("Normalize(%q) = %q,%v; want %q,%v", c.in, got, ok, c.want, c.wantOk)
		}
	}
}
