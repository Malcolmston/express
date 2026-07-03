package referer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func capture(t *testing.T, set func(*http.Request)) Referer {
	t.Helper()
	app := express.New()
	app.Use(New())
	var got Referer
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		got, _ = From(req)
		res.Send("ok")
	})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if set != nil {
		set(r)
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
	return got
}

func TestParsesHost(t *testing.T) {
	ref := capture(t, func(r *http.Request) {
		r.Header.Set("Referer", "https://example.com/page?x=1")
	})
	if ref.URL != "https://example.com/page?x=1" {
		t.Fatalf("url = %q", ref.URL)
	}
	if ref.Host != "example.com" {
		t.Fatalf("host = %q", ref.Host)
	}
}

func TestEmpty(t *testing.T) {
	ref := capture(t, nil)
	if ref.URL != "" || ref.Host != "" {
		t.Fatalf("got %+v", ref)
	}
}

func TestReferrerSpelling(t *testing.T) {
	ref := capture(t, func(r *http.Request) {
		r.Header.Set("Referrer", "https://alt.example.org/")
	})
	if ref.Host != "alt.example.org" {
		t.Fatalf("host = %q", ref.Host)
	}
}
