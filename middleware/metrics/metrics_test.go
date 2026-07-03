package metrics_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/metrics"
)

func TestCountsByClass(t *testing.T) {
	h, m := metrics.New()
	app := express.New()
	app.Use(h)
	app.Get("/ok", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })
	app.Get("/bad", func(req *express.Request, res *express.Response, next express.Next) { res.Status(400).Send("bad") })
	app.Get("/boom", func(req *express.Request, res *express.Response, next express.Next) { res.Status(500).Send("boom") })

	hit := func(path string) {
		r := httptest.NewRequest("GET", path, nil)
		app.ServeHTTP(httptest.NewRecorder(), r)
	}
	hit("/ok")
	hit("/ok")
	hit("/bad")
	hit("/boom")

	snap := m.Snapshot()
	if snap["total"] != 4 {
		t.Fatalf("expected total 4, got %d", snap["total"])
	}
	if snap["2xx"] != 2 {
		t.Fatalf("expected 2xx 2, got %d", snap["2xx"])
	}
	if snap["4xx"] != 1 {
		t.Fatalf("expected 4xx 1, got %d", snap["4xx"])
	}
	if snap["5xx"] != 1 {
		t.Fatalf("expected 5xx 1, got %d", snap["5xx"])
	}
}

func TestConcurrent(t *testing.T) {
	h, m := metrics.New()
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })

	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			r := httptest.NewRequest("GET", "/", nil)
			app.ServeHTTP(httptest.NewRecorder(), r)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	if m.Snapshot()["total"] != 50 {
		t.Fatalf("expected 50, got %d", m.Snapshot()["total"])
	}
}
