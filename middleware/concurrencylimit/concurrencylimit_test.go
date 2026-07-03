package concurrencylimit_test

import (
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/concurrencylimit"
)

func TestRejectsWhenFull(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{})
	app := express.New()
	app.Use(concurrencylimit.New(concurrencylimit.Options{Max: 1}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		entered <- struct{}{}
		<-release
		res.Send("ok")
	})

	// Occupy the single slot in a background goroutine.
	go func() {
		r := httptest.NewRequest("GET", "/", nil)
		app.ServeHTTP(httptest.NewRecorder(), r)
	}()
	<-entered // ensure the slot is held

	// Second request should be rejected immediately.
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 503 {
		t.Fatalf("expected 503, got %d", w.Code)
	}
	close(release)
}

func TestReleasesSlot(t *testing.T) {
	app := express.New()
	app.Use(concurrencylimit.New(concurrencylimit.Options{Max: 1}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	// Sequential requests should all succeed since the slot is released.
	for i := 0; i < 5; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("request %d expected 200, got %d", i, w.Code)
		}
	}
}

func TestAllowsUpToMaxConcurrent(t *testing.T) {
	release := make(chan struct{})
	var wg sync.WaitGroup
	entered := make(chan struct{}, 2)
	app := express.New()
	app.Use(concurrencylimit.New(concurrencylimit.Options{Max: 2}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		entered <- struct{}{}
		<-release
		res.Send("ok")
	})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, r)
			if w.Code != 200 {
				t.Errorf("expected 200, got %d", w.Code)
			}
		}()
	}
	<-entered
	<-entered // both slots occupied concurrently
	close(release)
	wg.Wait()
}
