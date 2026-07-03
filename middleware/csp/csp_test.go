package csp

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
	if got := rec.Header().Get("Content-Security-Policy"); got != "default-src 'self'" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildSortedAndJoined(t *testing.T) {
	rec := do(t, New(Options{Directives: map[string][]string{
		"script-src":  {"'self'", "https://cdn.example.com"},
		"default-src": {"'self'"},
	}}))
	want := "default-src 'self'; script-src 'self' https://cdn.example.com"
	if got := rec.Header().Get("Content-Security-Policy"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReportOnly(t *testing.T) {
	rec := do(t, New(Options{ReportOnly: true}))
	if got := rec.Header().Get("Content-Security-Policy-Report-Only"); got != "default-src 'self'" {
		t.Fatalf("got %q", got)
	}
	if got := rec.Header().Get("Content-Security-Policy"); got != "" {
		t.Fatalf("enforcing header should be empty, got %q", got)
	}
}

func TestDirectiveWithNoSources(t *testing.T) {
	got := Build(map[string][]string{"upgrade-insecure-requests": nil})
	if got != "upgrade-insecure-requests" {
		t.Fatalf("got %q", got)
	}
}
