package retryafter_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/retryafter"
)

func TestSetsHeaderOn503(t *testing.T) {
	app := express.New()
	app.Use(retryafter.New(retryafter.Options{Seconds: 30}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(503).Send("down")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Header().Get("Retry-After") != "30" {
		t.Fatalf("expected Retry-After 30, got %q", w.Header().Get("Retry-After"))
	}
}

func TestNoHeaderOn200(t *testing.T) {
	app := express.New()
	app.Use(retryafter.New(retryafter.Options{Seconds: 30}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Header().Get("Retry-After") != "" {
		t.Fatalf("did not expect Retry-After, got %q", w.Header().Get("Retry-After"))
	}
}

func TestDoesNotOverrideExisting(t *testing.T) {
	app := express.New()
	app.Use(retryafter.New(retryafter.Options{Seconds: 30}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Retry-After", "5")
		res.Status(429).Send("slow")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Header().Get("Retry-After") != "5" {
		t.Fatalf("expected existing 5 preserved, got %q", w.Header().Get("Retry-After"))
	}
}

func TestSetHelper(t *testing.T) {
	app := express.New()
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		retryafter.Set(res, 42)
		res.Status(503).Send("down")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Header().Get("Retry-After") != "42" {
		t.Fatalf("expected 42, got %q", w.Header().Get("Retry-After"))
	}
}
