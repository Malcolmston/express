package requestcounter_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/requestcounter"
)

func TestCounts(t *testing.T) {
	h, count := requestcounter.New()
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })

	for i := 0; i < 7; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		app.ServeHTTP(httptest.NewRecorder(), r)
	}
	if count() != 7 {
		t.Fatalf("expected 7, got %d", count())
	}
}

func TestConcurrent(t *testing.T) {
	h, count := requestcounter.New()
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })

	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			r := httptest.NewRequest("GET", "/", nil)
			app.ServeHTTP(httptest.NewRecorder(), r)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	if count() != 100 {
		t.Fatalf("expected 100, got %d", count())
	}
}
