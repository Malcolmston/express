package latencyheader_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/latencyheader"
)

func TestLatencyHeaderDeterministic(t *testing.T) {
	cur := time.Unix(0, 0)
	app := express.New()
	app.Use(latencyheader.New(latencyheader.Options{Now: func() time.Time { return cur }}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		cur = cur.Add(25 * time.Millisecond)
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if got := w.Header().Get("X-Response-Latency"); got != "25" {
		t.Fatalf("expected 25, got %q", got)
	}
}

func TestCustomHeader(t *testing.T) {
	cur := time.Unix(0, 0)
	app := express.New()
	app.Use(latencyheader.New(latencyheader.Options{Header: "X-Took", Now: func() time.Time { return cur }}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		cur = cur.Add(3 * time.Millisecond)
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if got := w.Header().Get("X-Took"); got != "3" {
		t.Fatalf("expected 3, got %q", got)
	}
}
