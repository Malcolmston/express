package timeout_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/timeout"
)

func TestFastHandlerPasses(t *testing.T) {
	app := express.New()
	app.Use(timeout.New(timeout.Options{Duration: 200 * time.Millisecond}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "ok" {
		t.Fatalf("expected 200 ok, got %d %q", w.Code, w.Body.String())
	}
}

func TestSlowHandlerTimesOut(t *testing.T) {
	app := express.New()
	app.Use(timeout.New(timeout.Options{Duration: 20 * time.Millisecond}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		time.Sleep(80 * time.Millisecond)
		res.Send("late")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 503 {
		t.Fatalf("expected 503, got %d", w.Code)
	}
	if w.Body.String() != "Service Unavailable" {
		t.Fatalf("unexpected body %q", w.Body.String())
	}
	// Give the late handler time to run; its write must be suppressed.
	time.Sleep(90 * time.Millisecond)
	if w.Body.String() != "Service Unavailable" {
		t.Fatalf("late write leaked into response: %q", w.Body.String())
	}
}

func TestCustomMessage(t *testing.T) {
	app := express.New()
	app.Use(timeout.New(timeout.Options{Duration: 10 * time.Millisecond, Message: "too slow"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		time.Sleep(60 * time.Millisecond)
		res.Send("late")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Body.String() != "too slow" {
		t.Fatalf("expected custom message, got %q", w.Body.String())
	}
}
