package timeout_test

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/timeout"
)

// ExampleNew demonstrates the timeout middleware converting a slow handler into
// a prompt 503 response. It mounts a timeout with a short 20ms Duration and a
// custom Message on an express.Application, then registers a "GET /" handler
// that deliberately sleeps well past the deadline before attempting to write.
// Driving the request through httptest, the deadline fires first, so the
// middleware wins the response race and sends 503 Service Unavailable with the
// configured message while the late handler's write is silently suppressed.
// The example sleeps briefly afterward to prove the background handler's late
// write never leaks into the already-sent response. Because the outcome here is
// governed by fixed, well-separated durations, the Output block is
// deterministic.
func ExampleNew() {
	app := express.New()
	app.Use(timeout.New(timeout.Options{
		Duration: 20 * time.Millisecond,
		Message:  "upstream too slow",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		time.Sleep(80 * time.Millisecond)
		res.Send("late answer")
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	// Let the losing handler finish; its write must not corrupt the response.
	time.Sleep(90 * time.Millisecond)
	fmt.Printf("%d %s\n", w.Code, w.Body.String())
	// Output: 503 upstream too slow
}

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
