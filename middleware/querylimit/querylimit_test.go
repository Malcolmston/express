package querylimit_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/querylimit"
)

func newApp() *express.Application {
	app := express.New()
	app.Use(querylimit.New(querylimit.Options{MaxLength: 10}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })
	return app
}

func TestWithinLimit(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("GET", "/?a=1", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestExceedsLimit(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("GET", "/?q="+strings.Repeat("x", 50), nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 414 {
		t.Fatalf("expected 414, got %d", w.Code)
	}
}

func TestBoundary(t *testing.T) {
	app := newApp()
	// RawQuery exactly 10 chars: "aaaa=bbbbb" (10) should pass.
	r := httptest.NewRequest("GET", "/?aaaa=bbbbb", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200 at boundary, got %d", w.Code)
	}
}
