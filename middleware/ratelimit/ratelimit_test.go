package ratelimit_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/ratelimit"
)

func newApp(o ratelimit.Options) *express.Application {
	app := express.New()
	app.Use(ratelimit.New(o))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func do(app *express.Application) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "1.2.3.4:5555"
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w
}

func TestAllowsUpToMax(t *testing.T) {
	app := newApp(ratelimit.Options{Max: 3, Window: time.Minute})
	for i := 0; i < 3; i++ {
		if w := do(app); w.Code != 200 {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
	}
}

func TestRejectsOverMax(t *testing.T) {
	app := newApp(ratelimit.Options{Max: 2, Window: time.Minute})
	do(app)
	do(app)
	w := do(app)
	if w.Code != 429 {
		t.Fatalf("expected 429, got %d", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatalf("missing Retry-After header")
	}
	if w.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("expected remaining 0, got %q", w.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestHeadersPresent(t *testing.T) {
	app := newApp(ratelimit.Options{Max: 5, Window: time.Minute})
	w := do(app)
	if w.Header().Get("X-RateLimit-Limit") != "5" {
		t.Fatalf("expected limit 5, got %q", w.Header().Get("X-RateLimit-Limit"))
	}
	if w.Header().Get("X-RateLimit-Remaining") != "4" {
		t.Fatalf("expected remaining 4, got %q", w.Header().Get("X-RateLimit-Remaining"))
	}
	if w.Header().Get("X-RateLimit-Reset") == "" {
		t.Fatalf("missing reset header")
	}
}

func TestWindowResetWithClock(t *testing.T) {
	base := time.Unix(1000, 0)
	cur := base
	app := newApp(ratelimit.Options{Max: 1, Window: time.Minute, Now: func() time.Time { return cur }})
	if w := do(app); w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w := do(app); w.Code != 429 {
		t.Fatalf("expected 429, got %d", w.Code)
	}
	cur = base.Add(2 * time.Minute)
	if w := do(app); w.Code != 200 {
		t.Fatalf("after window reset expected 200, got %d", w.Code)
	}
}

func TestPerKeyIsolation(t *testing.T) {
	app := express.New()
	app.Use(ratelimit.New(ratelimit.Options{Max: 1, Window: time.Minute}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })

	req := func(ip string) int {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = ip + ":9999"
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}
	if code := req("10.0.0.1"); code != 200 {
		t.Fatalf("first client expected 200, got %d", code)
	}
	if code := req("10.0.0.2"); code != 200 {
		t.Fatalf("second client expected 200, got %d", code)
	}
}
