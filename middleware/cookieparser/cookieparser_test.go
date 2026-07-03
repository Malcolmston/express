package cookieparser

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestParsesCookies(t *testing.T) {
	app := express.New()
	app.Use(New())

	var got map[string]string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		got = From(req)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "a", Value: "1"})
	req.AddCookie(&http.Cookie{Name: "b", Value: "hello world"})
	app.ServeHTTP(rec, req)

	if got["a"] != "1" {
		t.Fatalf("a = %q, want 1", got["a"])
	}
	if got["b"] != "hello world" {
		t.Fatalf("b = %q, want %q", got["b"], "hello world")
	}
}

func TestFromWithoutMiddleware(t *testing.T) {
	app := express.New()
	var got map[string]string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		got = From(req)
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got == nil || len(got) != 0 {
		t.Fatalf("expected empty non-nil map, got %v", got)
	}
}
