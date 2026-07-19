package negotiator

// Upstream-parity vectors transcribed verbatim from the jshttp/negotiator test
// suite (master branch):
//
//	https://raw.githubusercontent.com/jshttp/negotiator/master/test/mediaType.js
//	https://raw.githubusercontent.com/jshttp/negotiator/master/test/charset.js
//	https://raw.githubusercontent.com/jshttp/negotiator/master/test/encoding.js
//	https://raw.githubusercontent.com/jshttp/negotiator/master/test/language.js
//
// Each case pairs a request header value with the ordered preference the
// upstream library returns. The upstream singular methods return `undefined`
// when nothing is acceptable; this port returns "" instead, so undefined is
// encoded as "" below. The upstream `{ preferred: [...] }` developer-preference
// option has no equivalent in this port's API and is intentionally not covered
// here (see notes in the task summary).

import (
	"net/http"
	"reflect"
	"testing"
)

func neg(key, val string) *Negotiator {
	h := http.Header{}
	if val != "\x00" { // sentinel for "header absent"
		h.Set(key, val)
	}
	return New(h)
}

const absent = "\x00"

// ---------------------------------------------------------------------------
// Media types
// ---------------------------------------------------------------------------

func TestParityMediaTypeSingle(t *testing.T) {
	cases := []struct {
		accept string
		want   string
	}{
		{absent, "*/*"},
		{"*/*", "*/*"},
		{"application/json", "application/json"},
		{"application/json;q=0", ""},
		{"application/json;q=0.2, text/html", "text/html"},
		{"text/*", "text/*"},
		{"text/plain, application/json;q=0.5, text/html, */*;q=0.1", "text/plain"},
		{"text/plain, application/json;q=0.5, text/html, text/xml, text/yaml, text/javascript, text/csv, text/css, text/rtf, text/markdown, application/octet-stream;q=0.2, */*;q=0.1", "text/plain"},
	}
	for _, c := range cases {
		if got := neg("Accept", c.accept).MediaType(); got != c.want {
			t.Errorf("MediaType() accept=%q: got %q want %q", c.accept, got, c.want)
		}
	}
}

func TestParityMediaTypeSingleArray(t *testing.T) {
	cases := []struct {
		accept    string
		available []string
		want      string
	}{
		{absent, []string{"text/html"}, "text/html"},
		{absent, []string{"text/html", "application/json"}, "text/html"},
		{absent, []string{"application/json", "text/html"}, "application/json"},
		{"*/*", []string{"text/html"}, "text/html"},
		{"*/*", []string{"application/json", "text/html"}, "application/json"},
		{"application/json", []string{"application/JSON"}, "application/JSON"},
		{"application/json", []string{"text/html"}, ""},
		{"application/json", []string{"text/html", "application/json"}, "application/json"},
		{"application/json;q=0.2, text/html", []string{"application/json"}, "application/json"},
		{"application/json;q=0.2, text/html", []string{"application/json", "text/html"}, "text/html"},
		{"application/json;q=0.2, text/html", []string{"text/html", "application/json"}, "text/html"},
		{"text/*", []string{"application/json"}, ""},
		{"text/*", []string{"application/json", "text/html"}, "text/html"},
		{"text/*, text/plain;q=0", []string{"application/json", "text/html"}, "text/html"},
		{"text/plain, application/json;q=0.5, text/html, */*;q=0.1", []string{"application/json", "text/plain", "text/html"}, "text/plain"},
		{"text/plain, application/json;q=0.5, text/html, */*;q=0.1", []string{"image/jpeg", "text/html"}, "text/html"},
		{"text/plain, application/json;q=0.5, text/html, */*;q=0.1", []string{"image/jpeg", "image/gif"}, "image/jpeg"},
	}
	for _, c := range cases {
		if got := neg("Accept", c.accept).MediaType(c.available...); got != c.want {
			t.Errorf("MediaType(%v) accept=%q: got %q want %q", c.available, c.accept, got, c.want)
		}
	}
}

func TestParityMediaTypesPreferred(t *testing.T) {
	cases := []struct {
		accept string
		want   []string
	}{
		{absent, []string{"*/*"}},
		{"*/*", []string{"*/*"}},
		{"application/json", []string{"application/json"}},
		{"application/json;q=0", []string{}},
		{"application/json;q=0.2, text/html", []string{"text/html", "application/json"}},
		{"text/*", []string{"text/*"}},
		{"text/*, text/plain;q=0", []string{"text/*"}},
		{"text/html;LEVEL=1", []string{"text/html"}},
		{`text/html;foo="bar,text/css;";fizz="buzz,5", text/plain`, []string{"text/html", "text/plain"}},
		{"text/plain, application/json;q=0.5, text/html, */*;q=0.1", []string{"text/plain", "text/html", "application/json", "*/*"}},
		{"text/plain, application/json;q=0.5, text/html, text/xml, text/yaml, text/javascript, text/csv, text/css, text/rtf, text/markdown, application/octet-stream;q=0.2, */*;q=0.1",
			[]string{"text/plain", "text/html", "text/xml", "text/yaml", "text/javascript", "text/csv", "text/css", "text/rtf", "text/markdown", "application/json", "application/octet-stream", "*/*"}},
	}
	for _, c := range cases {
		got := neg("Accept", c.accept).MediaTypes()
		if !eq(got, c.want) {
			t.Errorf("MediaTypes() accept=%q: got %v want %v", c.accept, got, c.want)
		}
	}
}

func TestParityMediaTypesNegotiated(t *testing.T) {
	cases := []struct {
		accept    string
		available []string
		want      []string
	}{
		{absent, []string{"application/json", "text/plain"}, []string{"application/json", "text/plain"}},
		{"*/*", []string{"application/json", "text/plain"}, []string{"application/json", "text/plain"}},
		{"*/*;q=0.8, text/*, image/*",
			[]string{"application/json", "text/html", "text/plain", "text/xml", "application/xml", "image/gif", "image/jpeg", "image/png", "audio/mp3", "application/javascript", "text/javascript"},
			[]string{"text/html", "text/plain", "text/xml", "text/javascript", "image/gif", "image/jpeg", "image/png", "application/json", "application/xml", "audio/mp3", "application/javascript"}},
		{"application/json", []string{"application/json"}, []string{"application/json"}},
		{"application/json", []string{"application/JSON"}, []string{"application/JSON"}},
		{"application/json", []string{"text/html", "application/json"}, []string{"application/json"}},
		{"application/json", []string{"boom", "application/json"}, []string{"application/json"}},
		{"application/json;q=0", []string{"application/json"}, []string{}},
		{"application/json;q=0", []string{"application/json", "text/html", "image/jpeg"}, []string{}},
		{"application/json;q=0.2, text/html", []string{"application/json", "text/html"}, []string{"text/html", "application/json"}},
		{"application/json;q=0.9, text/html;q=0.8, application/json;q=0.7", []string{"text/html", "application/json"}, []string{"application/json", "text/html"}},
		{"application/json, */*;q=0.1", []string{"text/html", "application/json"}, []string{"application/json", "text/html"}},
		{`application/xhtml+xml;profile="http://www.wapforum.org/xhtml"`, []string{`application/xhtml+xml;profile="http://www.wapforum.org/xhtml"`}, []string{`application/xhtml+xml;profile="http://www.wapforum.org/xhtml"`}},
		{"text/*", []string{"text/html", "application/json", "text/plain"}, []string{"text/html", "text/plain"}},
		{"text/*, text/html;level", []string{"text/html"}, []string{"text/html"}},
		{"text/*, text/plain;q=0", []string{"text/html", "text/plain"}, []string{"text/html"}},
		{"text/*, text/plain;q=0.5", []string{"text/html", "text/plain", "text/xml"}, []string{"text/html", "text/xml", "text/plain"}},
		{"text/html;level=1", []string{"text/html;level=1"}, []string{"text/html;level=1"}},
		{"text/html;level=1", []string{"text/html;Level=1"}, []string{"text/html;Level=1"}},
		{"text/html;level=1", []string{"text/html;level=2"}, []string{}},
		{"text/html;level=1", []string{"text/html"}, []string{}},
		{"text/html;level=1", []string{"text/html;level=1;foo=bar"}, []string{"text/html;level=1;foo=bar"}},
		{"text/html;level=1;foo=bar", []string{"text/html;level=1"}, []string{}},
		{"text/html;level=1;foo=bar", []string{"text/html;level=1;foo=bar"}, []string{"text/html;level=1;foo=bar"}},
		{"text/html;level=1;foo=bar", []string{"text/html;foo=bar;level=1"}, []string{"text/html;foo=bar;level=1"}},
		{`text/html;level=1;foo="bar"`, []string{"text/html;level=1;foo=bar"}, []string{"text/html;level=1;foo=bar"}},
		{`text/html;level=1;foo="bar"`, []string{`text/html;level=1;foo="bar"`}, []string{`text/html;level=1;foo="bar"`}},
		{`text/html;foo=";level=2;"`, []string{"text/html;level=2"}, []string{}},
		{`text/html;foo=";level=2;"`, []string{`text/html;foo=";level=2;"`}, []string{`text/html;foo=";level=2;"`}},
		{"text/html;LEVEL=1", []string{"text/html;level=1"}, []string{"text/html;level=1"}},
		{"text/html;LEVEL=1", []string{"text/html;Level=1"}, []string{"text/html;Level=1"}},
		{"text/html;LEVEL=1;level=2", []string{"text/html;level=2"}, []string{"text/html;level=2"}},
		{"text/html;LEVEL=1;level=2", []string{"text/html;level=1"}, []string{}},
		{"text/html;level=2", []string{"text/html;level=1"}, []string{}},
		{"text/html;level=2, text/html", []string{"text/html", "text/html;level=2"}, []string{"text/html;level=2", "text/html"}},
		{"text/html;level=2;q=0.1, text/html", []string{"text/html;level=2", "text/html"}, []string{"text/html", "text/html;level=2"}},
		{"text/html;level=2;q=0.1;level=1", []string{"text/html;level=1"}, []string{}},
		{"text/html;level=2;q=0.1, text/html;level=1, text/html;q=0.5", []string{"text/html;level=1", "text/html;level=2", "text/html"}, []string{"text/html;level=1", "text/html", "text/html;level=2"}},
		{"text/plain, application/json;q=0.5, text/html, */*;q=0.1", []string{"text/html", "text/plain"}, []string{"text/plain", "text/html"}},
		{"text/plain, application/json;q=0.5, text/html, */*;q=0.1", []string{"application/json", "text/html", "text/plain"}, []string{"text/plain", "text/html", "application/json"}},
		{"text/plain, application/json;q=0.5, text/html, */*;q=0.1", []string{"image/jpeg", "text/html", "text/plain"}, []string{"text/plain", "text/html", "image/jpeg"}},
		{"text/plain, application/json;q=0.5, text/html, text/xml, text/yaml, text/javascript, text/csv, text/css, text/rtf, text/markdown, application/octet-stream;q=0.2, */*;q=0.1",
			[]string{"text/plain", "text/html", "text/xml", "text/yaml", "text/javascript", "text/csv", "text/css", "text/rtf", "text/markdown", "application/json", "application/octet-stream"},
			[]string{"text/plain", "text/html", "text/xml", "text/yaml", "text/javascript", "text/csv", "text/css", "text/rtf", "text/markdown", "application/json", "application/octet-stream"}},
	}
	for _, c := range cases {
		got := neg("Accept", c.accept).MediaTypes(c.available...)
		if !eq(got, c.want) {
			t.Errorf("MediaTypes(%v) accept=%q: got %v want %v", c.available, c.accept, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Charsets
// ---------------------------------------------------------------------------

func TestParityCharsetSingle(t *testing.T) {
	cases := []struct {
		accept string
		want   string
	}{
		{absent, "*"},
		{"*", "*"},
		{"*, UTF-8", "*"},
		{"*, UTF-8;q=0", "*"},
		{"ISO-8859-1", "ISO-8859-1"},
		{"UTF-8;q=0", ""},
		{"UTF-8, ISO-8859-1", "UTF-8"},
		{"UTF-8;q=0.8, ISO-8859-1", "ISO-8859-1"},
		{"UTF-8;q=0.9, ISO-8859-1;q=0.8, UTF-8;q=0.7", "UTF-8"},
	}
	for _, c := range cases {
		if got := neg("Accept-Charset", c.accept).Charset(); got != c.want {
			t.Errorf("Charset() accept=%q: got %q want %q", c.accept, got, c.want)
		}
	}
}

func TestParityCharsetSingleArray(t *testing.T) {
	cases := []struct {
		accept    string
		available []string
		want      string
	}{
		{absent, []string{"UTF-8"}, "UTF-8"},
		{absent, []string{"UTF-8", "ISO-8859-1"}, "UTF-8"},
		{"*", []string{"UTF-8"}, "UTF-8"},
		{"*, UTF-8", []string{"UTF-8", "ISO-8859-1"}, "UTF-8"},
		{"*, UTF-8;q=0", []string{"UTF-8", "ISO-8859-1"}, "ISO-8859-1"},
		{"*, UTF-8;q=0", []string{"UTF-8"}, ""},
		{"ISO-8859-1", []string{"ISO-8859-1"}, "ISO-8859-1"},
		{"ISO-8859-1", []string{"UTF-8", "ISO-8859-1"}, "ISO-8859-1"},
		{"ISO-8859-1", []string{"iso-8859-1"}, "iso-8859-1"},
		{"ISO-8859-1", []string{"iso-8859-1", "ISO-8859-1"}, "iso-8859-1"},
		{"ISO-8859-1", []string{"ISO-8859-1", "iso-8859-1"}, "ISO-8859-1"},
		{"ISO-8859-1", []string{"utf-8"}, ""},
		{"UTF-8;q=0", []string{"ISO-8859-1"}, ""},
		{"UTF-8;q=0", []string{"UTF-8", "KOI8-R", "ISO-8859-1"}, ""},
		{"UTF-8, ISO-8859-1", []string{"ISO-8859-1"}, "ISO-8859-1"},
		{"UTF-8, ISO-8859-1", []string{"UTF-8", "KOI8-R", "ISO-8859-1"}, "UTF-8"},
		{"UTF-8, ISO-8859-1", []string{"KOI8-R"}, ""},
		{"UTF-8;q=0.8, ISO-8859-1", []string{"ISO-8859-1"}, "ISO-8859-1"},
		{"UTF-8;q=0.8, ISO-8859-1", []string{"UTF-8", "KOI8-R", "ISO-8859-1"}, "ISO-8859-1"},
		{"UTF-8;q=0.8, ISO-8859-1", []string{"UTF-8", "KOI8-R"}, "UTF-8"},
		{"UTF-8;q=0.9, ISO-8859-1;q=0.8, UTF-8;q=0.7", []string{"UTF-8", "ISO-8859-1"}, "UTF-8"},
		{"UTF-8;q=0.9, ISO-8859-1;q=0.8, UTF-8;q=0.7", []string{"ISO-8859-1", "UTF-8"}, "UTF-8"},
	}
	for _, c := range cases {
		if got := neg("Accept-Charset", c.accept).Charset(c.available...); got != c.want {
			t.Errorf("Charset(%v) accept=%q: got %q want %q", c.available, c.accept, got, c.want)
		}
	}
}

func TestParityCharsetsPreferred(t *testing.T) {
	cases := []struct {
		accept string
		want   []string
	}{
		{absent, []string{"*"}},
		{"*", []string{"*"}},
		{"*, UTF-8", []string{"*", "UTF-8"}},
		{"*, UTF-8;q=0", []string{"*"}},
		{"UTF-8;q=0", []string{}},
		{"ISO-8859-1", []string{"ISO-8859-1"}},
		{"UTF-8, ISO-8859-1", []string{"UTF-8", "ISO-8859-1"}},
		{"UTF-8;q=0.8, ISO-8859-1", []string{"ISO-8859-1", "UTF-8"}},
		{"UTF-8;foo=bar;q=1, ISO-8859-1;q=1", []string{"UTF-8", "ISO-8859-1"}},
	}
	for _, c := range cases {
		got := neg("Accept-Charset", c.accept).Charsets()
		if !eq(got, c.want) {
			t.Errorf("Charsets() accept=%q: got %v want %v", c.accept, got, c.want)
		}
	}
}

func TestParityCharsetsNegotiated(t *testing.T) {
	cases := []struct {
		accept    string
		available []string
		want      []string
	}{
		{absent, []string{"UTF-8"}, []string{"UTF-8"}},
		{absent, []string{"UTF-8", "ISO-8859-1"}, []string{"UTF-8", "ISO-8859-1"}},
		{"*", []string{"UTF-8"}, []string{"UTF-8"}},
		{"*", []string{"UTF-8", "ISO-8859-1"}, []string{"UTF-8", "ISO-8859-1"}},
		{"*, UTF-8", []string{"UTF-8"}, []string{"UTF-8"}},
		{"*, UTF-8", []string{"UTF-8", "ISO-8859-1"}, []string{"UTF-8", "ISO-8859-1"}},
		{"*, UTF-8;q=0", []string{"UTF-8"}, []string{}},
		{"*, UTF-8;q=0", []string{"UTF-8", "ISO-8859-1"}, []string{"ISO-8859-1"}},
		{"UTF-8;q=0", []string{"ISO-8859-1"}, []string{}},
		{"UTF-8;q=0", []string{"UTF-8", "KOI8-R", "ISO-8859-1"}, []string{}},
		{"ISO-8859-1", []string{"ISO-8859-1"}, []string{"ISO-8859-1"}},
		{"ISO-8859-1", []string{"UTF-8", "ISO-8859-1"}, []string{"ISO-8859-1"}},
		{"ISO-8859-1", []string{"iso-8859-1"}, []string{"iso-8859-1"}},
		{"ISO-8859-1", []string{"iso-8859-1", "ISO-8859-1"}, []string{"iso-8859-1", "ISO-8859-1"}},
		{"ISO-8859-1", []string{"ISO-8859-1", "iso-8859-1"}, []string{"ISO-8859-1", "iso-8859-1"}},
		{"ISO-8859-1", []string{"utf-8"}, []string{}},
		{"UTF-8, ISO-8859-1", []string{"ISO-8859-1"}, []string{"ISO-8859-1"}},
		{"UTF-8, ISO-8859-1", []string{"UTF-8", "KOI8-R", "ISO-8859-1"}, []string{"UTF-8", "ISO-8859-1"}},
		{"UTF-8, ISO-8859-1", []string{"KOI8-R"}, []string{}},
		{"UTF-8;q=0.8, ISO-8859-1", []string{"ISO-8859-1"}, []string{"ISO-8859-1"}},
		{"UTF-8;q=0.8, ISO-8859-1", []string{"UTF-8", "KOI8-R", "ISO-8859-1"}, []string{"ISO-8859-1", "UTF-8"}},
		{"UTF-8;q=0.9, ISO-8859-1;q=0.8, UTF-8;q=0.7", []string{"UTF-8", "ISO-8859-1"}, []string{"UTF-8", "ISO-8859-1"}},
		{"UTF-8;q=0.9, ISO-8859-1;q=0.8, UTF-8;q=0.7", []string{"ISO-8859-1", "UTF-8"}, []string{"UTF-8", "ISO-8859-1"}},
	}
	for _, c := range cases {
		got := neg("Accept-Charset", c.accept).Charsets(c.available...)
		if !eq(got, c.want) {
			t.Errorf("Charsets(%v) accept=%q: got %v want %v", c.available, c.accept, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Encodings
// ---------------------------------------------------------------------------

func TestParityEncodingSingle(t *testing.T) {
	cases := []struct {
		accept string
		want   string
	}{
		{absent, "identity"},
		{"*", "*"},
		{"*, gzip", "*"},
		{"*, gzip;q=0", "*"},
		{"*;q=0", ""},
		{"*;q=0, identity;q=1", "identity"},
		{"identity", "identity"},
		{"identity;q=0", ""},
		{"gzip", "gzip"},
		{"gzip, compress;q=0", "gzip"},
		{"gzip, deflate", "gzip"},
		{"gzip;q=0.8, deflate", "deflate"},
		{"gzip;q=0.8, identity;q=0.5, *;q=0.3", "gzip"},
	}
	for _, c := range cases {
		if got := neg("Accept-Encoding", c.accept).Encoding(); got != c.want {
			t.Errorf("Encoding() accept=%q: got %q want %q", c.accept, got, c.want)
		}
	}
}

func TestParityEncodingSingleArray(t *testing.T) {
	cases := []struct {
		accept    string
		available []string
		want      string
	}{
		{absent, []string{"identity"}, "identity"},
		{absent, []string{"gzip"}, ""},
		{"*", []string{"identity"}, "identity"},
		{"*", []string{"gzip"}, "gzip"},
		{"*", []string{"gzip", "identity"}, "gzip"},
		{"*, gzip", []string{"identity"}, "identity"},
		{"*, gzip", []string{"gzip"}, "gzip"},
		{"*, gzip", []string{"compress", "gzip"}, "gzip"},
		{"*, gzip;q=0", []string{"identity"}, "identity"},
		{"*, gzip;q=0", []string{"gzip"}, ""},
		{"*, gzip;q=0", []string{"gzip", "compress"}, "compress"},
		{"*;q=0", []string{"identity"}, ""},
		{"*;q=0", []string{"gzip"}, ""},
		{"*;q=0, identity;q=1", []string{"identity"}, "identity"},
		{"*;q=0, identity;q=1", []string{"gzip"}, ""},
		{"identity", []string{"identity"}, "identity"},
		{"identity", []string{"gzip"}, ""},
		{"identity;q=0", []string{"identity"}, ""},
		{"identity;q=0", []string{"gzip"}, ""},
		{"gzip", []string{"gzip"}, "gzip"},
		{"gzip", []string{"identity", "gzip"}, "gzip"},
		{"gzip", []string{"identity"}, "identity"},
		{"gzip, compress;q=0", []string{"compress"}, ""},
		{"gzip, compress;q=0", []string{"deflate", "compress"}, ""},
		{"gzip, compress;q=0", []string{"gzip", "compress"}, "gzip"},
		{"gzip, deflate", []string{"deflate", "compress"}, "deflate"},
		{"gzip;q=0.8, deflate", []string{"gzip"}, "gzip"},
		{"gzip;q=0.8, deflate", []string{"deflate"}, "deflate"},
		{"gzip;q=0.8, deflate", []string{"deflate", "gzip"}, "deflate"},
		{"gzip;q=0.8, identity;q=0.5, *;q=0.3", []string{"gzip"}, "gzip"},
		{"gzip;q=0.8, identity;q=0.5, *;q=0.3", []string{"compress", "identity"}, "identity"},
	}
	for _, c := range cases {
		if got := neg("Accept-Encoding", c.accept).Encoding(c.available...); got != c.want {
			t.Errorf("Encoding(%v) accept=%q: got %q want %q", c.available, c.accept, got, c.want)
		}
	}
}

func TestParityEncodingsPreferred(t *testing.T) {
	cases := []struct {
		accept string
		want   []string
	}{
		{absent, []string{"identity"}},
		{"*", []string{"*"}},
		{"*, gzip", []string{"*", "gzip"}},
		{"*, gzip;q=0", []string{"*"}},
		{"*;q=0", []string{}},
		{"*;q=0, identity;q=1", []string{"identity"}},
		{"identity", []string{"identity"}},
		{"identity;q=0", []string{}},
		{"gzip", []string{"gzip", "identity"}},
		{"gzip, compress;q=0", []string{"gzip", "identity"}},
		{"gzip, deflate", []string{"gzip", "deflate", "identity"}},
		{"gzip;q=0.8, deflate", []string{"deflate", "gzip", "identity"}},
		{"gzip;foo=bar;q=1, deflate;q=1", []string{"gzip", "deflate", "identity"}},
		{"gzip;q=0.8, identity;q=0.5, *;q=0.3", []string{"gzip", "identity", "*"}},
	}
	for _, c := range cases {
		got := neg("Accept-Encoding", c.accept).Encodings()
		if !eq(got, c.want) {
			t.Errorf("Encodings() accept=%q: got %v want %v", c.accept, got, c.want)
		}
	}
}

func TestParityEncodingsNegotiated(t *testing.T) {
	cases := []struct {
		accept    string
		available []string
		want      []string
	}{
		{absent, []string{"identity"}, []string{"identity"}},
		{absent, []string{"gzip"}, []string{}},
		{"*", []string{"identity"}, []string{"identity"}},
		{"*", []string{"gzip"}, []string{"gzip"}},
		{"*", []string{"gzip", "identity"}, []string{"gzip", "identity"}},
		{"*, gzip", []string{"identity"}, []string{"identity"}},
		{"*, gzip", []string{"gzip"}, []string{"gzip"}},
		{"*, gzip", []string{"compress", "gzip"}, []string{"gzip", "compress"}},
		{"*, gzip;q=0", []string{"identity"}, []string{"identity"}},
		{"*, gzip;q=0", []string{"gzip"}, []string{}},
		{"*, gzip;q=0", []string{"gzip", "compress"}, []string{"compress"}},
		{"*;q=0", []string{"identity"}, []string{}},
		{"*;q=0", []string{"gzip"}, []string{}},
		{"*;q=0, identity;q=1", []string{"identity"}, []string{"identity"}},
		{"*;q=0, identity;q=1", []string{"gzip"}, []string{}},
		{"identity", []string{"identity"}, []string{"identity"}},
		{"identity", []string{"gzip"}, []string{}},
		{"identity;q=0", []string{"identity"}, []string{}},
		{"identity;q=0", []string{"gzip"}, []string{}},
		{"gzip", []string{"GZIP"}, []string{"GZIP"}},
		{"gzip", []string{"gzip", "GZIP"}, []string{"gzip", "GZIP"}},
		{"gzip", []string{"GZIP", "gzip"}, []string{"GZIP", "gzip"}},
		{"gzip", []string{"gzip"}, []string{"gzip"}},
		{"gzip", []string{"gzip", "identity"}, []string{"gzip", "identity"}},
		{"gzip", []string{"identity", "gzip"}, []string{"gzip", "identity"}},
		{"gzip", []string{"identity"}, []string{"identity"}},
		{"gzip, compress;q=0", []string{"gzip", "compress"}, []string{"gzip"}},
		{"gzip, deflate", []string{"gzip"}, []string{"gzip"}},
		{"gzip, deflate", []string{"gzip", "identity"}, []string{"gzip", "identity"}},
		{"gzip, deflate", []string{"deflate", "gzip"}, []string{"gzip", "deflate"}},
		{"gzip, deflate", []string{"identity"}, []string{"identity"}},
		{"gzip;q=0.8, deflate", []string{"gzip"}, []string{"gzip"}},
		{"gzip;q=0.8, deflate", []string{"deflate"}, []string{"deflate"}},
		{"gzip;q=0.8, deflate", []string{"deflate", "gzip"}, []string{"deflate", "gzip"}},
		{"gzip;q=0.8, identity;q=0.5, *;q=0.3", []string{"gzip"}, []string{"gzip"}},
		{"gzip;q=0.8, identity;q=0.5, *;q=0.3", []string{"identity", "gzip", "compress"}, []string{"gzip", "identity", "compress"}},
	}
	for _, c := range cases {
		got := neg("Accept-Encoding", c.accept).Encodings(c.available...)
		if !eq(got, c.want) {
			t.Errorf("Encodings(%v) accept=%q: got %v want %v", c.available, c.accept, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Languages
// ---------------------------------------------------------------------------

func TestParityLanguageSingle(t *testing.T) {
	cases := []struct {
		accept string
		want   string
	}{
		{absent, "*"},
		{"*", "*"},
		{"*, en", "*"},
		{"*, en;q=0", "*"},
		{"*;q=0.8, en, es", "en"},
		{"en", "en"},
		{"en;q=0", ""},
		{"en;q=0.8, es", "es"},
		{"en;q=0.9, es;q=0.8, en;q=0.7", "en"},
		{"en-US, en;q=0.8", "en-US"},
		{"en-US, en-GB", "en-US"},
		{"en-US;q=0.8, es", "es"},
		{"nl;q=0.5, fr, de, en, it, es, pt, no, se, fi, ro", "fr"},
	}
	for _, c := range cases {
		if got := neg("Accept-Language", c.accept).Language(); got != c.want {
			t.Errorf("Language() accept=%q: got %q want %q", c.accept, got, c.want)
		}
	}
}

func TestParityLanguageSingleArray(t *testing.T) {
	cases := []struct {
		accept    string
		available []string
		want      string
	}{
		{absent, []string{"en"}, "en"},
		{absent, []string{"es", "en"}, "es"},
		{"*", []string{"en"}, "en"},
		{"*", []string{"es", "en"}, "es"},
		{"*, en", []string{"en"}, "en"},
		{"*, en", []string{"es", "en"}, "en"},
		{"*, en;q=0", []string{"en"}, ""},
		{"*, en;q=0", []string{"es", "en"}, "es"},
		{"*;q=0.8, en, es", []string{"en", "nl"}, "en"},
		{"*;q=0.8, en, es", []string{"ro", "nl"}, "ro"},
		{"en", []string{"en"}, "en"},
		{"en", []string{"es", "en"}, "en"},
		{"en", []string{"en-US"}, "en-US"},
		{"en", []string{"en-US", "en"}, "en"},
		{"en", []string{"en", "en-US"}, "en"},
		{"en;q=0", []string{"es", "en"}, ""},
		{"en;q=0.8, es", []string{"en"}, "en"},
		{"en;q=0.8, es", []string{"en", "es"}, "es"},
		{"en;q=0.9, es;q=0.8, en;q=0.7", []string{"es"}, "es"},
		{"en;q=0.9, es;q=0.8, en;q=0.7", []string{"en", "es"}, "en"},
		{"en;q=0.9, es;q=0.8, en;q=0.7", []string{"es", "en"}, "en"},
		{"en-US, en;q=0.8", []string{"en", "en-US"}, "en-US"},
		{"en-US, en;q=0.8", []string{"en-GB", "en-US"}, "en-US"},
		{"en-US, en;q=0.8", []string{"en-GB", "es"}, "en-GB"},
		{"en-US, en-GB", []string{"en-US", "en-GB"}, "en-US"},
		{"en-US, en-GB", []string{"en-GB", "en-US"}, "en-US"},
		{"en-US;q=0.8, es", []string{"es", "en-US"}, "es"},
		{"en-US;q=0.8, es", []string{"en-US", "es"}, "es"},
		{"en-US;q=0.8, es", []string{"en-US", "en"}, "en-US"},
		{"nl;q=0.5, fr, de, en, it, es, pt, no, se, fi, ro", []string{"nl", "fr"}, "fr"},
	}
	for _, c := range cases {
		if got := neg("Accept-Language", c.accept).Language(c.available...); got != c.want {
			t.Errorf("Language(%v) accept=%q: got %q want %q", c.available, c.accept, got, c.want)
		}
	}
}

func TestParityLanguagesPreferred(t *testing.T) {
	cases := []struct {
		accept string
		want   []string
	}{
		{absent, []string{"*"}},
		{"*", []string{"*"}},
		{"*, en", []string{"*", "en"}},
		{"*, en;q=0", []string{"*"}},
		{"*;q=0.8, en, es", []string{"en", "es", "*"}},
		{"en", []string{"en"}},
		{"en;q=0", []string{}},
		{"en;q=0.8, es", []string{"es", "en"}},
		{"en-US, en;q=0.8", []string{"en-US", "en"}},
		{"en-US, en-GB", []string{"en-US", "en-GB"}},
		{"en-US;q=0.8, es", []string{"es", "en-US"}},
		{"en-US;foo=bar;q=1, en-GB;q=1", []string{"en-US", "en-GB"}},
		{"nl;q=0.5, fr, de, en, it, es, pt, no, se, fi, ro", []string{"fr", "de", "en", "it", "es", "pt", "no", "se", "fi", "ro", "nl"}},
	}
	for _, c := range cases {
		got := neg("Accept-Language", c.accept).Languages()
		if !eq(got, c.want) {
			t.Errorf("Languages() accept=%q: got %v want %v", c.accept, got, c.want)
		}
	}
}

func TestParityLanguagesNegotiated(t *testing.T) {
	cases := []struct {
		accept    string
		available []string
		want      []string
	}{
		{absent, []string{"en"}, []string{"en"}},
		{absent, []string{"es", "en"}, []string{"es", "en"}},
		{"*", []string{"en"}, []string{"en"}},
		{"*", []string{"es", "en"}, []string{"es", "en"}},
		{"*, en", []string{"en"}, []string{"en"}},
		{"*, en", []string{"es", "en"}, []string{"en", "es"}},
		{"*, en;q=0", []string{"en"}, []string{}},
		{"*, en;q=0", []string{"es", "en"}, []string{"es"}},
		{"*;q=0.8, en, es", []string{"fr", "de", "en", "it", "es", "pt", "no", "se", "fi", "ro", "nl"}, []string{"en", "es", "fr", "de", "it", "pt", "no", "se", "fi", "ro", "nl"}},
		{"en", []string{"en"}, []string{"en"}},
		{"en", []string{"en", "es"}, []string{"en"}},
		{"en", []string{"es", "en"}, []string{"en"}},
		{"en", []string{"en-US"}, []string{"en-US"}},
		{"en", []string{"en-US", "en"}, []string{"en", "en-US"}},
		{"en", []string{"en", "en-US"}, []string{"en", "en-US"}},
		{"en;q=0", []string{"en"}, []string{}},
		{"en;q=0", []string{"en", "es"}, []string{}},
		{"en;q=0.8, es", []string{"en"}, []string{"en"}},
		{"en;q=0.8, es", []string{"en", "es"}, []string{"es", "en"}},
		{"en;q=0.8, es", []string{"es", "en"}, []string{"es", "en"}},
		{"en-US, en;q=0.8", []string{"en-us", "EN"}, []string{"en-us", "EN"}},
		{"en-US, en;q=0.8", []string{"en-US", "en"}, []string{"en-US", "en"}},
		{"en-US, en;q=0.8", []string{"en-GB", "en-US", "en"}, []string{"en-US", "en", "en-GB"}},
		{"en-US, en-GB", []string{"en-US", "en-GB"}, []string{"en-US", "en-GB"}},
		{"en-US, en-GB", []string{"en-GB", "en-US"}, []string{"en-US", "en-GB"}},
		{"en-US;q=0.8, es", []string{"en", "es"}, []string{"es", "en"}},
		{"en-US;q=0.8, es", []string{"en", "es", "en-US"}, []string{"es", "en-US", "en"}},
		{"nl;q=0.5, fr, de, en, it, es, pt, no, se, fi, ro", []string{"fr", "de", "en", "it", "es", "pt", "no", "se", "fi", "ro", "nl"}, []string{"fr", "de", "en", "it", "es", "pt", "no", "se", "fi", "ro", "nl"}},
	}
	for _, c := range cases {
		got := neg("Accept-Language", c.accept).Languages(c.available...)
		if !eq(got, c.want) {
			t.Errorf("Languages(%v) accept=%q: got %v want %v", c.available, c.accept, got, c.want)
		}
	}
}

// eq compares two string slices, treating nil and empty as equal.
func eq(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}
