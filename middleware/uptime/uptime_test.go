package uptime_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/uptime"
)

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
