package cachecontrol

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func run(t *testing.T, o Options) string {
	t.Helper()
	app := express.New()
	app.Use(New(o))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	return rr.Header().Get("Cache-Control")
}

func TestCacheControlPublicMaxAge(t *testing.T) {
	if got := run(t, Options{Public: true, MaxAge: 3600}); got != "public, max-age=3600" {
		t.Fatalf("got %q", got)
	}
}

func TestCacheControlPrivateNoStore(t *testing.T) {
	if got := run(t, Options{Private: true, NoStore: true}); got != "private, no-store" {
		t.Fatalf("got %q", got)
	}
}

func TestCacheControlEmpty(t *testing.T) {
	if got := run(t, Options{}); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}
