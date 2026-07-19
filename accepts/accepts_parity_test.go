package accepts

// Upstream-parity tests for the Go port of jshttp/accepts.
//
// Every input -> expected-output vector below is transcribed directly from the
// original library's own mocha test suite (npm "accepts" v2.0.0), fetched from:
//
//   https://raw.githubusercontent.com/jshttp/accepts/master/test/type.js
//   https://raw.githubusercontent.com/jshttp/accepts/master/test/charset.js
//   https://raw.githubusercontent.com/jshttp/accepts/master/test/encoding.js
//   https://raw.githubusercontent.com/jshttp/accepts/master/test/language.js
//
// Mapping between the two APIs:
//   - Upstream accept.types()/charsets()/languages()/encodings() with NO
//     arguments return an ordered array -> Go Types()/Charsets()/Languages()/
//     Encodings() (plural, variadic with no offers).
//   - Upstream accept.types(a, b, ...) with arguments return the single best
//     match (a string) or false -> Go Type()/Charset()/Language()/Encoding()
//     (singular); upstream `false` maps to the Go empty string "".
//   - An upstream request created with no header argument (undefined header)
//     maps to an http.Header with the field unset; an empty-string header maps
//     to the field set to "".

import (
	"net/http"
	"reflect"
	"testing"
)

// header presence helpers ---------------------------------------------------

func hdrUnset() http.Header { return http.Header{} }

func hdrSet(key, val string) http.Header {
	h := http.Header{}
	h.Set(key, val)
	return h
}

// ---------------------------------------------------------------------------
// test/type.js
// ---------------------------------------------------------------------------

func TestParityTypesNoArguments(t *testing.T) {
	// Accept populated -> all accepted types in preference order.
	a := New(hdrSet("Accept", "application/*;q=0.2, image/jpeg;q=0.8, text/html, text/plain"))
	want := []string{"text/html", "text/plain", "image/jpeg", "application/*"}
	if got := a.Types(); !reflect.DeepEqual(got, want) {
		t.Errorf("Types() = %v, want %v", got, want)
	}

	// Accept not in request -> */*.
	if got := New(hdrUnset()).Types(); !reflect.DeepEqual(got, []string{"*/*"}) {
		t.Errorf("Types() absent = %v, want [*/*]", got)
	}

	// Accept empty -> [].
	if got := New(hdrSet("Accept", "")).Types(); len(got) != 0 {
		t.Errorf("Types() empty = %v, want []", got)
	}
}

func TestParityTypesNoValidTypes(t *testing.T) {
	a := New(hdrSet("Accept", "application/*;q=0.2, image/jpeg;q=0.8, text/html, text/plain"))
	if got := a.Type("image/png", "image/tiff"); got != "" {
		t.Errorf("Type(image/png, image/tiff) = %q, want \"\" (false)", got)
	}

	// Accept not populated -> first offer.
	if got := New(hdrUnset()).Type("text/html", "text/plain", "image/jpeg", "application/*"); got != "text/html" {
		t.Errorf("Type(...) absent = %q, want text/html", got)
	}
}

func TestParityTypesExtensions(t *testing.T) {
	a := New(hdrSet("Accept", "text/plain, text/html"))
	cases := []struct {
		offer string
		want  string
	}{
		{"html", "html"},
		{".html", ".html"},
		{"txt", "txt"},
		{".txt", ".txt"},
		{"png", ""},
		{"bogus", ""},
	}
	for _, c := range cases {
		if got := a.Type(c.offer); got != c.want {
			t.Errorf("Type(%q) = %q, want %q", c.offer, got, c.want)
		}
	}
}

func TestParityTypesFirstMatch(t *testing.T) {
	a := New(hdrSet("Accept", "text/plain, text/html"))
	cases := []struct {
		offers []string
		want   string
	}{
		{[]string{"png", "text", "html"}, "text"},
		{[]string{"png", "html"}, "html"},
		{[]string{"bogus", "html"}, "html"},
	}
	for _, c := range cases {
		if got := a.Type(c.offers...); got != c.want {
			t.Errorf("Type(%v) = %q, want %q", c.offers, got, c.want)
		}
	}
}

func TestParityTypesExactMatch(t *testing.T) {
	a := New(hdrSet("Accept", "text/plain, text/html"))
	if got := a.Type("text/html"); got != "text/html" {
		t.Errorf("Type(text/html) = %q", got)
	}
	if got := a.Type("text/plain"); got != "text/plain" {
		t.Errorf("Type(text/plain) = %q", got)
	}
}

func TestParityTypesTypeMatch(t *testing.T) {
	a := New(hdrSet("Accept", "application/json, */*"))
	for _, offer := range []string{"text/html", "text/plain", "image/png"} {
		if got := a.Type(offer); got != offer {
			t.Errorf("Type(%q) = %q, want %q", offer, got, offer)
		}
	}
}

func TestParityTypesSubtypeMatch(t *testing.T) {
	a := New(hdrSet("Accept", "application/json, text/*"))
	if got := a.Type("text/html"); got != "text/html" {
		t.Errorf("Type(text/html) = %q", got)
	}
	if got := a.Type("text/plain"); got != "text/plain" {
		t.Errorf("Type(text/plain) = %q", got)
	}
	if got := a.Type("image/png"); got != "" {
		t.Errorf("Type(image/png) = %q, want \"\" (false)", got)
	}
}

// ---------------------------------------------------------------------------
// test/charset.js
// ---------------------------------------------------------------------------

func TestParityCharsetsNoArguments(t *testing.T) {
	a := New(hdrSet("Accept-Charset", "utf-8, iso-8859-1;q=0.2, utf-7;q=0.5"))
	want := []string{"utf-8", "utf-7", "iso-8859-1"}
	if got := a.Charsets(); !reflect.DeepEqual(got, want) {
		t.Errorf("Charsets() = %v, want %v", got, want)
	}

	if got := New(hdrUnset()).Charsets(); !reflect.DeepEqual(got, []string{"*"}) {
		t.Errorf("Charsets() absent = %v, want [*]", got)
	}

	if got := New(hdrSet("Accept-Charset", "")).Charsets(); len(got) != 0 {
		t.Errorf("Charsets() empty = %v, want []", got)
	}
}

func TestParityCharsetsBestFit(t *testing.T) {
	a := New(hdrSet("Accept-Charset", "utf-8, iso-8859-1;q=0.2, utf-7;q=0.5"))
	if got := a.Charset("utf-7", "utf-8"); got != "utf-8" {
		t.Errorf("Charset(utf-7, utf-8) = %q, want utf-8", got)
	}
	if got := a.Charset("utf-16"); got != "" {
		t.Errorf("Charset(utf-16) = %q, want \"\" (false)", got)
	}
}

func TestParityCharsetsNotPopulated(t *testing.T) {
	if got := New(hdrUnset()).Charset("utf-7", "utf-8"); got != "utf-7" {
		t.Errorf("Charset(utf-7, utf-8) absent = %q, want utf-7", got)
	}
}

// ---------------------------------------------------------------------------
// test/encoding.js
// ---------------------------------------------------------------------------

func TestParityEncodingsNoArguments(t *testing.T) {
	a := New(hdrSet("Accept-Encoding", "gzip, compress;q=0.2"))
	want := []string{"gzip", "compress", "identity"}
	if got := a.Encodings(); !reflect.DeepEqual(got, want) {
		t.Errorf("Encodings() = %v, want %v", got, want)
	}
	if got := a.Encoding("gzip", "compress"); got != "gzip" {
		t.Errorf("Encoding(gzip, compress) = %q, want gzip", got)
	}
}

func TestParityEncodingsAbsentHeader(t *testing.T) {
	a := New(hdrUnset())
	if got := a.Encodings(); !reflect.DeepEqual(got, []string{"identity"}) {
		t.Errorf("Encodings() absent = %v, want [identity]", got)
	}
	if got := a.Encoding("gzip", "deflate", "identity"); got != "identity" {
		t.Errorf("Encoding(gzip, deflate, identity) absent = %q, want identity", got)
	}
	if got := a.Encoding("gzip", "deflate"); got != "" {
		t.Errorf("Encoding(gzip, deflate) absent = %q, want \"\" (false)", got)
	}
}

func TestParityEncodingsEmptyHeader(t *testing.T) {
	a := New(hdrSet("Accept-Encoding", ""))
	if got := a.Encodings(); !reflect.DeepEqual(got, []string{"identity"}) {
		t.Errorf("Encodings() empty = %v, want [identity]", got)
	}
	if got := a.Encoding("gzip", "deflate", "identity"); got != "identity" {
		t.Errorf("Encoding(gzip, deflate, identity) empty = %q, want identity", got)
	}
	if got := a.Encoding("gzip", "deflate"); got != "" {
		t.Errorf("Encoding(gzip, deflate) empty = %q, want \"\" (false)", got)
	}
}

func TestParityEncodingsBestFit(t *testing.T) {
	a := New(hdrSet("Accept-Encoding", "gzip, compress;q=0.2"))
	if got := a.Encoding("compress", "gzip"); got != "gzip" {
		t.Errorf("Encoding(compress, gzip) = %q, want gzip", got)
	}
	if got := a.Encoding("gzip", "compress"); got != "gzip" {
		t.Errorf("Encoding(gzip, compress) = %q, want gzip", got)
	}
}

// ---------------------------------------------------------------------------
// test/language.js
// ---------------------------------------------------------------------------

func TestParityLanguagesNoArguments(t *testing.T) {
	a := New(hdrSet("Accept-Language", "en;q=0.8, es, pt"))
	want := []string{"es", "pt", "en"}
	if got := a.Languages(); !reflect.DeepEqual(got, want) {
		t.Errorf("Languages() = %v, want %v", got, want)
	}

	if got := New(hdrUnset()).Languages(); !reflect.DeepEqual(got, []string{"*"}) {
		t.Errorf("Languages() absent = %v, want [*]", got)
	}

	if got := New(hdrSet("Accept-Language", "")).Languages(); len(got) != 0 {
		t.Errorf("Languages() empty = %v, want []", got)
	}
}

func TestParityLanguagesBestFit(t *testing.T) {
	a := New(hdrSet("Accept-Language", "en;q=0.8, es, pt"))
	if got := a.Language("es", "en"); got != "es" {
		t.Errorf("Language(es, en) = %q, want es", got)
	}
	if got := a.Language("fr", "au"); got != "" {
		t.Errorf("Language(fr, au) = %q, want \"\" (false)", got)
	}
}

func TestParityLanguagesNotPopulated(t *testing.T) {
	if got := New(hdrUnset()).Language("es", "en"); got != "es" {
		t.Errorf("Language(es, en) absent = %q, want es", got)
	}
}
