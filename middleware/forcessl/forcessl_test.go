package forcessl

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func do(t *testing.T, h express.Handler, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	app := express.New()
	app.Use(h)
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("secure-ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, r)
	return rec
}

func TestRedirectsInsecure(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "http://example.com/path?x=1", nil)
	rec := do(t, New(), r)
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("code = %d, want 301", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com/path?x=1" {
		t.Fatalf("location = %q", loc)
	}
}

func TestForwardedProtoHTTPS(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	r.Header.Set("X-Forwarded-Proto", "https")
	rec := do(t, New(), r)
	if rec.Body.String() != "secure-ok" {
		t.Fatalf("expected pass-through, got %q", rec.Body.String())
	}
}

func TestDisabled(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rec := do(t, New(Options{Enabled: false}), r)
	if rec.Body.String() != "secure-ok" {
		t.Fatalf("expected pass-through when disabled, got %q", rec.Body.String())
	}
}
