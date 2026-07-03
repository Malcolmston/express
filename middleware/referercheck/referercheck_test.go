package referercheck_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/referercheck"
)

func newApp(optional bool) *express.Application {
	app := express.New()
	app.Use(referercheck.New(referercheck.Options{
		Allow:    []string{"example.com"},
		Optional: optional,
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func run(app *express.Application, ref string) int {
	r := httptest.NewRequest("GET", "/", nil)
	if ref != "" {
		r.Header.Set("Referer", ref)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w.Code
}

func TestAllowedReferer(t *testing.T) {
	if code := run(newApp(false), "https://example.com/page"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestDisallowedReferer(t *testing.T) {
	if code := run(newApp(false), "https://evil.com/page"); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestMissingRefererRejected(t *testing.T) {
	if code := run(newApp(false), ""); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestMissingRefererOptional(t *testing.T) {
	if code := run(newApp(true), ""); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}
