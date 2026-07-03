package requestdump

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestCapturesLast(t *testing.T) {
	Reset()
	app := express.New()
	app.Use(New())
	app.Get("/foo", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set("X-Test", "yes")
	app.ServeHTTP(rec, req)

	d := Last()
	if d.Method != http.MethodGet || d.Path != "/foo" {
		t.Fatalf("got %+v", d)
	}
	if d.Headers.Get("X-Test") != "yes" {
		t.Fatalf("headers not captured: %v", d.Headers)
	}
}

func TestRingBounded(t *testing.T) {
	Reset()
	app := express.New()
	app.Use(New(Options{Size: 3}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	if got := len(All()); got != 3 {
		t.Fatalf("ring size = %d, want 3", got)
	}
}
