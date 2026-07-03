package maintenance_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/maintenance"
)

func TestToggle(t *testing.T) {
	h, toggle := maintenance.New(maintenance.Options{RetryAfter: 120})
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })

	// Initially off.
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200 when off, got %d", w.Code)
	}

	// Turn on.
	toggle.Set(true)
	if !toggle.Enabled() {
		t.Fatalf("expected Enabled true")
	}
	r = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 503 {
		t.Fatalf("expected 503 when on, got %d", w.Code)
	}
	if w.Header().Get("Retry-After") != "120" {
		t.Fatalf("expected Retry-After 120, got %q", w.Header().Get("Retry-After"))
	}

	// Turn off again.
	toggle.Set(false)
	r = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200 after disabling, got %d", w.Code)
	}
}

func TestEnabledFunc(t *testing.T) {
	on := true
	h, _ := maintenance.New(maintenance.Options{EnabledFunc: func() bool { return on }})
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 503 {
		t.Fatalf("expected 503, got %d", w.Code)
	}
	on = false
	r = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
