package slowdown_test

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/slowdown"
)

func TestNoDelayUnderThreshold(t *testing.T) {
	var mu sync.Mutex
	var delays []time.Duration
	app := express.New()
	app.Use(slowdown.New(slowdown.Options{
		Threshold: 3,
		Delay:     10 * time.Millisecond,
		Sleep:     func(d time.Duration) { mu.Lock(); delays = append(delays, d); mu.Unlock() },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })

	for i := 0; i < 3; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "1.1.1.1:1"
		app.ServeHTTP(httptest.NewRecorder(), r)
	}
	if len(delays) != 0 {
		t.Fatalf("expected no delays under threshold, got %v", delays)
	}
}

func TestIncreasingDelay(t *testing.T) {
	var got []time.Duration
	app := express.New()
	app.Use(slowdown.New(slowdown.Options{
		Threshold: 2,
		Delay:     10 * time.Millisecond,
		Sleep:     func(d time.Duration) { got = append(got, d) },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		d, _ := req.Value("slowdown-delay")
		res.Send("ok")
		_ = d
	})

	for i := 0; i < 5; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "2.2.2.2:1"
		app.ServeHTTP(httptest.NewRecorder(), r)
	}
	// Requests 3,4,5 exceed threshold => delays of 1x,2x,3x.
	want := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond}
	if len(got) != len(want) {
		t.Fatalf("expected %d delays, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("delay %d: expected %v, got %v", i, want[i], got[i])
		}
	}
}

func TestMaxDelayCap(t *testing.T) {
	var got []time.Duration
	app := express.New()
	app.Use(slowdown.New(slowdown.Options{
		Threshold: 1,
		Delay:     50 * time.Millisecond,
		MaxDelay:  60 * time.Millisecond,
		Sleep:     func(d time.Duration) { got = append(got, d) },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("ok") })

	for i := 0; i < 4; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "3.3.3.3:1"
		app.ServeHTTP(httptest.NewRecorder(), r)
	}
	for _, d := range got {
		if d > 60*time.Millisecond {
			t.Fatalf("delay %v exceeds cap", d)
		}
	}
}
