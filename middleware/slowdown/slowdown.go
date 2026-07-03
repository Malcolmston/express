// Package slowdown provides middleware that progressively delays responses for
// a client once it exceeds a configured request threshold within a fixed
// window. It mirrors the behaviour of express-slow-down: the first requests are
// served immediately, and each subsequent request beyond the threshold incurs
// an increasing delay.
//
// So that it remains deterministic and testable, the actual sleeping is
// delegated to a Sleep hook (defaulting to time.Sleep). The computed delay is
// also stored on the request via req.Set("slowdown-delay", d) for inspection.
package slowdown

import (
	"sync"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the slow-down middleware.
type Options struct {
	// Threshold is the number of requests allowed per window before delays are
	// applied. Values <= 0 default to 5.
	Threshold int
	// Window is the length of the fixed counting window. Values <= 0 default to
	// one minute.
	Window time.Duration
	// Delay is the incremental delay added per request beyond the threshold.
	// Values <= 0 default to 100ms.
	Delay time.Duration
	// MaxDelay caps the per-request delay. A value <= 0 means no cap.
	MaxDelay time.Duration
	// KeyFunc derives the bucket key. When nil the client IP is used.
	KeyFunc func(req *express.Request) string
	// Sleep performs the delay. When nil it defaults to time.Sleep. Tests may
	// inject a no-op or recording function.
	Sleep func(time.Duration)
	// Now returns the current time. When nil it defaults to time.Now.
	Now func() time.Time
}

type window struct {
	count int
	reset time.Time
}

// New returns slow-down middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Threshold <= 0 {
		o.Threshold = 5
	}
	if o.Window <= 0 {
		o.Window = time.Minute
	}
	if o.Delay <= 0 {
		o.Delay = 100 * time.Millisecond
	}
	if o.KeyFunc == nil {
		o.KeyFunc = func(req *express.Request) string { return req.IP() }
	}
	if o.Sleep == nil {
		o.Sleep = time.Sleep
	}
	now := o.Now
	if now == nil {
		now = time.Now
	}

	var mu sync.Mutex
	buckets := make(map[string]*window)

	return func(req *express.Request, res *express.Response, next express.Next) {
		key := o.KeyFunc(req)
		t := now()

		mu.Lock()
		w, ok := buckets[key]
		if !ok || !t.Before(w.reset) {
			w = &window{reset: t.Add(o.Window)}
			buckets[key] = w
		}
		w.count++
		count := w.count
		mu.Unlock()

		var delay time.Duration
		if count > o.Threshold {
			delay = time.Duration(count-o.Threshold) * o.Delay
			if o.MaxDelay > 0 && delay > o.MaxDelay {
				delay = o.MaxDelay
			}
		}
		req.Set("slowdown-delay", delay)
		if delay > 0 {
			o.Sleep(delay)
		}
		next()
	}
}
