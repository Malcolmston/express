package negotiator

import (
	"net/http"
	"reflect"
	"testing"
)

func hdr(kv map[string]string) http.Header {
	h := http.Header{}
	for k, v := range kv {
		h.Set(k, v)
	}
	return h
}

func TestMediaTypeQValueOrdering(t *testing.T) {
	n := New(hdr(map[string]string{"Accept": "text/html;q=0.8, application/json;q=0.9"}))
	got := n.MediaTypes("text/html", "application/json")
	want := []string{"application/json", "text/html"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if mt := n.MediaType("text/html", "application/json"); mt != "application/json" {
		t.Fatalf("MediaType got %q", mt)
	}
}

func TestMediaTypeWildcard(t *testing.T) {
	n := New(hdr(map[string]string{"Accept": "*/*"}))
	if mt := n.MediaType("text/html"); mt != "text/html" {
		t.Fatalf("*/* should match, got %q", mt)
	}
	n2 := New(hdr(map[string]string{"Accept": "text/*"}))
	if mt := n2.MediaType("text/plain"); mt != "text/plain" {
		t.Fatalf("text/* should match text/plain, got %q", mt)
	}
	if mt := n2.MediaType("application/json"); mt != "" {
		t.Fatalf("text/* should not match application/json, got %q", mt)
	}
}

func TestMediaTypeNoHeader(t *testing.T) {
	n := New(http.Header{})
	if mt := n.MediaType("text/html"); mt != "text/html" {
		t.Fatalf("missing Accept should accept all, got %q", mt)
	}
}

func TestMediaTypesAll(t *testing.T) {
	n := New(hdr(map[string]string{"Accept": "text/html;q=0.5, application/json"}))
	got := n.MediaTypes()
	want := []string{"application/json", "text/html"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestLanguagePrefix(t *testing.T) {
	n := New(hdr(map[string]string{"Accept-Language": "en"}))
	if l := n.Language("en-US", "fr"); l != "en-US" {
		t.Fatalf("en should match en-US, got %q", l)
	}
}

func TestLanguageQValue(t *testing.T) {
	n := New(hdr(map[string]string{"Accept-Language": "en;q=0.5, fr;q=0.9"}))
	if l := n.Language("en", "fr"); l != "fr" {
		t.Fatalf("got %q want fr", l)
	}
	got := n.Languages("en", "fr")
	want := []string{"fr", "en"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestCharset(t *testing.T) {
	n := New(hdr(map[string]string{"Accept-Charset": "utf-8, iso-8859-1;q=0.5"}))
	if c := n.Charset("iso-8859-1", "utf-8"); c != "utf-8" {
		t.Fatalf("got %q want utf-8", c)
	}
}

func TestEncodingIdentity(t *testing.T) {
	// identity is acceptable even when not listed.
	n := New(hdr(map[string]string{"Accept-Encoding": "gzip"}))
	if e := n.Encoding("identity"); e != "identity" {
		t.Fatalf("identity should be acceptable, got %q", e)
	}
	if e := n.Encoding("gzip", "identity"); e != "gzip" {
		t.Fatalf("gzip preferred, got %q", e)
	}
}

func TestEncodingIdentityDisabled(t *testing.T) {
	n := New(hdr(map[string]string{"Accept-Encoding": "gzip, identity;q=0"}))
	if e := n.Encoding("identity"); e != "" {
		t.Fatalf("identity disabled should not be acceptable, got %q", e)
	}
}

func TestEncodingOrdering(t *testing.T) {
	n := New(hdr(map[string]string{"Accept-Encoding": "gzip;q=0.8, br;q=0.9"}))
	got := n.Encodings("gzip", "br")
	want := []string{"br", "gzip"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
