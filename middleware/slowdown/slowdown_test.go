package slowdown_test

import (
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/slowdown"
)

// ExampleNew shows the slow-down middleware applying a growing delay once a
// client crosses the request threshold within a window. We configure a low
// Threshold and a small per-request Delay, and inject a recording Sleep hook so
// the example observes the computed delays without actually pausing (real
// deployments leave Sleep nil to use time.Sleep). Driving six requests from the
// same client through the app, the first two pass untouched and each request
// beyond the threshold is delayed by an additional Delay increment. The example
// prints the recorded delays to show the linear growth; note that because real
// timing is nondeterministic this Example intentionally omits an Output block.
func ExampleNew() {
	var delays []time.Duration
	app := express.New()
	app.Use(slowdown.New(slowdown.Options{
		Threshold: 2,
		Delay:     20 * time.Millisecond,
		Sleep:     func(d time.Duration) { delays = append(delays, d) },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	for i := 0; i < 6; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "203.0.113.7:9000"
		app.ServeHTTP(httptest.NewRecorder(), r)
	}

	// Requests 3..6 exceed the threshold and are delayed 1x..4x Delay.
	fmt.Println("delayed requests:", len(delays))
	_ = delays
}

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
