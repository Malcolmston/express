package vhost

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func do(t *testing.T, h express.Handler, host string) string {
	t.Helper()
	app := express.New()
	app.Use(h)
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("fallthrough")
	})
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://"+host+"/", nil)
	app.ServeHTTP(rec, r)
	return rec.Body.String()
}

func apiHandler(req *express.Request, res *express.Response, next express.Next) {
	res.Send("api")
}

func TestExactMatch(t *testing.T) {
	h := New(Options{Host: "api.example.com", Handler: apiHandler})
	if got := do(t, h, "api.example.com"); got != "api" {
		t.Fatalf("got %q", got)
	}
}

func TestNoMatch(t *testing.T) {
	h := New(Options{Host: "api.example.com", Handler: apiHandler})
	if got := do(t, h, "www.example.com"); got != "fallthrough" {
		t.Fatalf("got %q", got)
	}
}

func TestWildcardMatch(t *testing.T) {
	h := New(Options{Host: "*.example.com", Handler: apiHandler})
	if got := do(t, h, "foo.example.com"); got != "api" {
		t.Fatalf("got %q", got)
	}
}

func TestWildcardExcludesBare(t *testing.T) {
	h := New(Options{Host: "*.example.com", Handler: apiHandler})
	if got := do(t, h, "example.com"); got != "fallthrough" {
		t.Fatalf("got %q", got)
	}
}
