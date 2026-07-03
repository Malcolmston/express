package methodoverride

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func run(t *testing.T, h express.Handler, r *http.Request) string {
	t.Helper()
	app := express.New()
	app.Use(h)
	var seen string
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		seen = req.Method()
		res.Send("ok")
	})
	app.ServeHTTP(httptest.NewRecorder(), r)
	return seen
}

func TestHeaderOverride(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set(DefaultHeader, "delete")
	if got := run(t, New(), r); got != "DELETE" {
		t.Fatalf("got %q, want DELETE", got)
	}
}

func TestQueryOverride(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/?_method=put", nil)
	if got := run(t, New(), r); got != "PUT" {
		t.Fatalf("got %q, want PUT", got)
	}
}

func TestHeaderBeatsQuery(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/?_method=put", nil)
	r.Header.Set(DefaultHeader, "patch")
	if got := run(t, New(), r); got != "PATCH" {
		t.Fatalf("got %q, want PATCH", got)
	}
}

func TestNotPost(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?_method=delete", nil)
	if got := run(t, New(), r); got != "GET" {
		t.Fatalf("got %q, want GET (no override on GET)", got)
	}
}

func TestCustomOptions(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/?verb=put", nil)
	if got := run(t, New(Options{Query: "verb"}), r); got != "PUT" {
		t.Fatalf("got %q, want PUT", got)
	}
}
