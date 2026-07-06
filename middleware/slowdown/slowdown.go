// Package slowdown provides middleware that progressively delays responses for
// a client once it exceeds a configured request threshold within a fixed
// window. It is the express framework's Go analogue of the Node
// express-slow-down package (express-rate-limit/express-slow-down): rather than
// rejecting bursts outright the way a rate limiter does, it lets the first
// requests through untouched and then adds a growing pause to each further
// request, gently discouraging abuse while keeping legitimate traffic working.
//
// Reach for this middleware to blunt brute-force and scraping attempts, to
// smooth spikes against an expensive endpoint, or as a softer companion to a
// hard rate limiter — the slow-down absorbs the shoulder of a burst so fewer
// requests ever reach the limiter's cutoff. It pairs naturally with a login or
// search route where a slightly slower response is an acceptable price for
// throttling an attacker but a limiter's outright 429 would be too blunt.
//
// Operationally the middleware sits near the front of the chain. On each request
// it derives a bucket key with Options.KeyFunc (the client IP via req.IP by
// default), looks up that key's fixed window, and either reuses the current
// window or, if the window's reset time has passed, starts a fresh one of length
// Options.Window. It increments the window's count under a mutex, and when the
// count exceeds Options.Threshold it computes a delay of (count-Threshold) *
// Options.Delay, capped at Options.MaxDelay when that is set. The delay is
// recorded on the request with req.Set("slowdown-delay", d) for inspection or
// logging, the Options.Sleep hook is invoked to actually wait, and then next()
// is always called — this middleware never short-circuits or rejects a request,
// it only slows it.
//
// Several defaults and semantics matter. Threshold defaults to 5, Window to one
// minute, and Delay to 100ms; a MaxDelay of 0 means the delay grows without
// bound. Counting uses a fixed window (not a sliding one), so all buckets tied
// to a key reset together when the window elapses, which permits a brief burst
// at a window boundary. The Sleep and Now hooks default to time.Sleep and
// time.Now and exist chiefly for testing: injecting a recording Sleep and a
// controllable Now makes the otherwise time-dependent behavior deterministic.
// The in-memory bucket map is guarded by a sync.Mutex, so the middleware is safe
// for concurrent requests within a single process.
//
// Compared with express-slow-down this port keeps the core "delay after N
// requests, increasing per request" contract but is deliberately compact. Its
// counters live only in process memory, so they are per-instance and lost on
// restart with no Redis or shared-store backend; it exposes no
// skipSuccessfulRequests/skipFailedRequests filtering, no per-key limit
// overrides, and no response headers advertising the current delay. The delay is
// applied by sleeping before next() rather than by scheduling, and the growth is
// strictly linear in the request count with an optional hard cap.
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
