package throttle_test

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/throttle"
)

// ExampleNew demonstrates the token-bucket rate limiter guarding a route. It
// builds a throttle with Rate 1 and Burst 2 driven by a fixed clock, mounts it
// on an express.Application ahead of a "GET /" handler, and then drives three
// back-to-back requests from the same client through httptest. The first two
// requests fit within the burst and succeed, while the third finds an empty
// bucket and is short-circuited with 429 Too Many Requests plus a Retry-After
// header. The clock is frozen so no tokens refill between requests, making the
// rejection point deterministic. Output is intentionally omitted from the doc
// example because throttle timing is nondeterministic in general use.
func ExampleNew() {
	clock := time.Unix(0, 0)
	app := express.New()
	app.Use(throttle.New(throttle.Options{
		Rate:  1,
		Burst: 2,
		Now:   func() time.Time { return clock },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	for i := 0; i < 3; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "203.0.113.7:5000"
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		fmt.Printf("request %d -> %d retry-after=%q\n", i+1, w.Code, w.Header().Get("Retry-After"))
	}
}

func newApp(o throttle.Options) *express.Application {
	app := express.New()
	app.Use(throttle.New(o))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })
	return app
}

func do(app *express.Application) int {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "9.9.9.9:1"
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w.Code
}

func TestBurstThenReject(t *testing.T) {
	cur := time.Unix(0, 0)
	app := newApp(throttle.Options{Rate: 1, Burst: 2, Now: func() time.Time { return cur }})
	if code := do(app); code != 200 {
		t.Fatalf("req1 expected 200, got %d", code)
	}
	if code := do(app); code != 200 {
		t.Fatalf("req2 expected 200, got %d", code)
	}
	if code := do(app); code != 429 {
		t.Fatalf("req3 expected 429, got %d", code)
	}
}

func TestRefillOverTime(t *testing.T) {
	cur := time.Unix(0, 0)
	app := newApp(throttle.Options{Rate: 1, Burst: 1, Now: func() time.Time { return cur }})
	if code := do(app); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	if code := do(app); code != 429 {
		t.Fatalf("expected 429, got %d", code)
	}
	// Advance one second -> one token refilled.
	cur = cur.Add(time.Second)
	if code := do(app); code != 200 {
		t.Fatalf("after refill expected 200, got %d", code)
	}
}

func TestRetryAfterHeader(t *testing.T) {
	cur := time.Unix(0, 0)
	app := newApp(throttle.Options{Rate: 0.5, Burst: 1, Now: func() time.Time { return cur }})
	do(app)
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "9.9.9.9:1"
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 429 {
		t.Fatalf("expected 429, got %d", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatalf("missing Retry-After")
	}
}
