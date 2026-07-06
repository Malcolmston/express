package uptime_test

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/uptime"
)

// ExampleNew demonstrates stamping an uptime header and reading the same value
// programmatically. It constructs the middleware with an injected clock so the
// example is repeatable, capturing both the express.Handler and the since
// accessor that New returns as a pair. The handler is mounted on an
// express.Application in front of a "GET /" route; after advancing the clock by
// 42 seconds the request is driven through httptest, and the middleware writes
// the elapsed whole seconds to the X-Uptime header before calling next(). The
// since accessor reports the identical elapsed duration at full precision,
// showing how application code can surface uptime without parsing the header.
// The Output block would be nondeterministic under a real wall clock, so an
// injected clock is used to keep it stable.
func ExampleNew() {
	clock := time.Unix(1000, 0)
	handler, since := uptime.New(uptime.Options{
		Now: func() time.Time { return clock },
	})

	app := express.New()
	app.Use(handler)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	clock = clock.Add(42 * time.Second)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Printf("X-Uptime=%s since=%v\n", w.Header().Get("X-Uptime"), since())
}

func TestUptimeHeader(t *testing.T) {
	cur := time.Unix(1000, 0)
	h, since := uptime.New(uptime.Options{Now: func() time.Time { return cur }})
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })

	cur = cur.Add(42 * time.Second)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if got := w.Header().Get("X-Uptime"); got != "42" {
		t.Fatalf("expected 42, got %q", got)
	}
	if since() != 42*time.Second {
		t.Fatalf("expected Since 42s, got %v", since())
	}
}

func TestCustomHeader(t *testing.T) {
	cur := time.Unix(0, 0)
	h, _ := uptime.New(uptime.Options{Header: "X-Up", Now: func() time.Time { return cur }})
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })
	cur = cur.Add(5 * time.Second)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if got := w.Header().Get("X-Up"); got != "5" {
		t.Fatalf("expected 5, got %q", got)
	}
}
