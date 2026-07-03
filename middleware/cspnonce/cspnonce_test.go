package cspnonce

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

func TestSetsCSPWithNonce(t *testing.T) {
	app := express.New()
	app.Use(New())
	var n string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		n = Nonce(req)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if n == "" {
		t.Fatalf("expected a nonce")
	}
	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src") {
		t.Fatalf("CSP missing script-src: %q", csp)
	}
	if !strings.Contains(csp, "'nonce-"+n+"'") {
		t.Fatalf("CSP %q does not contain nonce %q", csp, n)
	}
}

func TestCustomDirectives(t *testing.T) {
	app := express.New()
	app.Use(New(Options{DefaultSrc: "'none'", ScriptSrc: "https://cdn.example.com"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.HasPrefix(csp, "default-src 'none';") {
		t.Fatalf("CSP = %q", csp)
	}
	if !strings.Contains(csp, "https://cdn.example.com") {
		t.Fatalf("CSP = %q", csp)
	}
}
