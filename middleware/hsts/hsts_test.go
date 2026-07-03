package hsts

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func do(t *testing.T, h express.Handler) *httptest.ResponseRecorder {
	t.Helper()
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec
}

func TestDefault(t *testing.T) {
	rec := do(t, New())
	if got := rec.Header().Get("Strict-Transport-Security"); got != "max-age=15552000" {
		t.Fatalf("got %q", got)
	}
}

func TestFullOptions(t *testing.T) {
	rec := do(t, New(Options{MaxAge: 100, IncludeSubDomains: true, Preload: true}))
	want := "max-age=100; includeSubDomains; preload"
	if got := rec.Header().Get("Strict-Transport-Security"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNegativeMaxAge(t *testing.T) {
	rec := do(t, New(Options{MaxAge: -1}))
	if got := rec.Header().Get("Strict-Transport-Security"); got != "max-age=0" {
		t.Fatalf("got %q", got)
	}
}
