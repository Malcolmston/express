package healthz_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/healthz"
)

func TestHealthzResponds(t *testing.T) {
	app := express.New()
	app.Use(healthz.New())
	r := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "ok" {
		t.Fatalf("expected 200 ok, got %d %q", w.Code, w.Body.String())
	}
}

func TestFallThrough(t *testing.T) {
	app := express.New()
	app.Use(healthz.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("home") })
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Body.String() != "home" {
		t.Fatalf("expected fall-through, got %q", w.Body.String())
	}
}

func TestCustomPath(t *testing.T) {
	app := express.New()
	app.Use(healthz.New(healthz.Options{Path: "/live", Body: "alive"}))
	r := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "alive" {
		t.Fatalf("expected alive, got %d %q", w.Code, w.Body.String())
	}
}
