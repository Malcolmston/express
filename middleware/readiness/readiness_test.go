package readiness_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/readiness"
)

func TestReady(t *testing.T) {
	app := express.New()
	app.Use(readiness.New(readiness.Options{Ready: func() bool { return true }}))
	r := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestNotReady(t *testing.T) {
	ready := false
	app := express.New()
	app.Use(readiness.New(readiness.Options{Ready: func() bool { return ready }}))
	r := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 503 {
		t.Fatalf("expected 503, got %d", w.Code)
	}
	ready = true
	r = httptest.NewRequest("GET", "/readyz", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200 after ready, got %d", w.Code)
	}
}

func TestFallThrough(t *testing.T) {
	app := express.New()
	app.Use(readiness.New(readiness.Options{Path: "/readyz"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("home") })
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Body.String() != "home" {
		t.Fatalf("expected fall-through, got %q", w.Body.String())
	}
}
