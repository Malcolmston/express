package accepts

import (
	"net/http"
	"testing"
)

func makeHeader(kv map[string]string) http.Header {
	h := http.Header{}
	for k, v := range kv {
		h.Set(k, v)
	}
	return h
}

func TestTypeBasic(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept": "text/html"}))
	if got := a.Type("html"); got != "html" {
		t.Errorf("Type(html) = %q", got)
	}
	if got := a.Type("json"); got != "" {
		t.Errorf("Type(json) = %q, want empty", got)
	}
}

func TestTypeQValues(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept": "text/html;q=0.5, application/json;q=0.9"}))
	if got := a.Type("html", "json"); got != "json" {
		t.Errorf("Type = %q, want json", got)
	}
}

func TestTypeWildcard(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept": "*/*"}))
	if got := a.Type("json"); got != "json" {
		t.Errorf("Type = %q, want json", got)
	}
}

func TestTypeTextWildcard(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept": "text/*"}))
	if got := a.Type("html"); got != "html" {
		t.Errorf("Type(html) = %q", got)
	}
	if got := a.Type("json"); got != "" {
		t.Errorf("Type(json) = %q, want empty", got)
	}
}

func TestTypeNoHeader(t *testing.T) {
	a := New(http.Header{})
	if got := a.Type("json", "html"); got != "json" {
		t.Errorf("Type = %q, want json (first offer)", got)
	}
}

func TestTypeFullType(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept": "application/json"}))
	if got := a.Type("application/json"); got != "application/json" {
		t.Errorf("Type = %q", got)
	}
}

func TestTypesNoOffers(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept": "text/html, application/json;q=0.9"}))
	got := a.Types()
	if len(got) != 2 || got[0] != "text/html" {
		t.Errorf("Types() = %v", got)
	}
}

func TestTypesExcludesQZero(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept": "text/html, application/json;q=0"}))
	if got := a.Type("json"); got != "" {
		t.Errorf("Type(json) = %q, want empty (q=0)", got)
	}
}

func TestLanguage(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept-Language": "en-US, en;q=0.8, fr;q=0.5"}))
	if got := a.Language("en", "fr"); got != "en" {
		t.Errorf("Language = %q, want en", got)
	}
	if got := a.Language("fr", "de"); got != "fr" {
		t.Errorf("Language = %q, want fr", got)
	}
	if got := a.Language("de"); got != "" {
		t.Errorf("Language(de) = %q, want empty", got)
	}
}

func TestLanguageWildcard(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept-Language": "*"}))
	if got := a.Language("en", "fr"); got != "en" {
		t.Errorf("Language = %q", got)
	}
}

func TestLanguagesNoOffers(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept-Language": "en;q=0.8, fr"}))
	got := a.Languages()
	if len(got) != 2 || got[0] != "fr" {
		t.Errorf("Languages() = %v, want fr first", got)
	}
}

func TestCharset(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept-Charset": "utf-8, iso-8859-1;q=0.5"}))
	if got := a.Charset("utf-8", "iso-8859-1"); got != "utf-8" {
		t.Errorf("Charset = %q", got)
	}
	if got := a.Charset("iso-8859-1"); got != "iso-8859-1" {
		t.Errorf("Charset = %q", got)
	}
	if got := a.Charset("us-ascii"); got != "" {
		t.Errorf("Charset(us-ascii) = %q, want empty", got)
	}
}

func TestCharsetWildcard(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept-Charset": "*"}))
	if got := a.Charset("utf-8"); got != "utf-8" {
		t.Errorf("Charset = %q", got)
	}
}

func TestEncoding(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept-Encoding": "gzip, deflate"}))
	if got := a.Encoding("gzip", "deflate"); got != "gzip" {
		t.Errorf("Encoding = %q", got)
	}
	if got := a.Encoding("br"); got != "" {
		t.Errorf("Encoding(br) = %q, want empty", got)
	}
}

func TestEncodingIdentityAlwaysAcceptable(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept-Encoding": "gzip"}))
	if got := a.Encoding("identity"); got != "identity" {
		t.Errorf("Encoding(identity) = %q, want identity", got)
	}
}

func TestEncodingIdentityDisallowed(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept-Encoding": "gzip, identity;q=0"}))
	if got := a.Encoding("identity"); got != "" {
		t.Errorf("Encoding(identity) = %q, want empty (disallowed)", got)
	}
}

func TestEncodingQValues(t *testing.T) {
	a := New(makeHeader(map[string]string{"Accept-Encoding": "gzip;q=0.5, br;q=1.0"}))
	if got := a.Encoding("gzip", "br"); got != "br" {
		t.Errorf("Encoding = %q, want br", got)
	}
}

func TestEncodingsNoHeaderIdentity(t *testing.T) {
	a := New(http.Header{})
	got := a.Encoding("identity")
	if got != "identity" {
		t.Errorf("Encoding(identity) with no header = %q", got)
	}
}
