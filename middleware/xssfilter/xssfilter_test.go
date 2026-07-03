package xssfilter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func do(t *testing.T, h express.Handler) string {
	t.Helper()
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec.Header().Get("X-XSS-Protection")
}

func TestDefault(t *testing.T) {
	if got := do(t, New()); got != "0" {
		t.Fatalf("got %q", got)
	}
}

func TestCustom(t *testing.T) {
	if got := do(t, New(Options{Value: "1; mode=block"})); got != "1; mode=block" {
		t.Fatalf("got %q", got)
	}
}
