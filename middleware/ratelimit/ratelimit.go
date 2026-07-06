// Package ratelimit provides a fixed-window, per-client rate limiter for the
// express framework. Each client (keyed by IP by default) is allowed a maximum
// number of requests within a fixed window; requests beyond the limit receive a
// 429 Too Many Requests response carrying standard rate-limit headers.
//
// It ports the fixed-window strategy popularized by the Node express-rate-limit
// middleware: a counter per key that resets at a fixed boundary, together with
// the conventional X-RateLimit-Limit, X-RateLimit-Remaining, and
// X-RateLimit-Reset response headers (and a Retry-After header on rejection).
// Like the classic express-rate-limit default store, counters are held in an
// in-process map, so the limiter is self-contained and requires no external
// datastore.
//
// Use it to protect endpoints from abuse, brute-force attempts, or accidental
// request floods when you want a simple, dependency-free throttle. Register it
// with app.Use ahead of the handlers you want to guard (globally, or on a
// specific sub-path) so the counter is checked before the protected work runs.
// Because it is a leading middleware, it can short-circuit the chain and answer
// with 429 without ever invoking the route handler.
//
// On each request the middleware derives a bucket key via KeyFunc (the client
// IP from req.IP by default) and reads the current time via Now. Under a mutex
// it looks up the key's window: if none exists or the window has expired
// (the current time is at or after the window's reset instant) a fresh window is
// started with reset = now + Window and count 0. The count is incremented, then
// the response always receives X-RateLimit-Limit (Max), X-RateLimit-Remaining
// (Max - count, floored at 0), and X-RateLimit-Reset (the reset time as a Unix
// timestamp). When the post-increment count is within Max the middleware calls
// next() to continue; when it exceeds Max it additionally sets Retry-After (the
// whole seconds until reset, at least 1) and responds 429 with the body
// "Too Many Requests", returning without calling next().
//
// The window is fixed, not sliding: all requests that fall before a window's
// reset instant share one counter that snaps back to zero only when that
// instant passes, which permits a short burst around a window boundary (up to
// roughly 2*Max across two adjacent windows) — the well-known trade-off of the
// fixed-window algorithm. Options carry the tunables and their defaults: Max
// defaults to 60 when <= 0, Window defaults to one minute when <= 0, KeyFunc
// defaults to keying by req.IP when nil, and Now defaults to time.Now when nil.
// Injecting a fixed Now is the supported way to make behaviour deterministic in
// tests. The counter map grows one entry per distinct key and entries are not
// actively evicted, so a custom KeyFunc over an unbounded key space (or a very
// large client population) can accumulate memory; keying by IP as in the default
// keeps this bounded in practice.
//
// Parity with the Node original is close in shape but intentionally minimal:
// the fixed window, the per-key counter, the 429 status, and the standard
// X-RateLimit-* and Retry-After headers all match express-rate-limit's legacy
// header mode, while advanced features of the Node package — pluggable external
// stores (Redis, Memcached), the draft standard "RateLimit" combined header,
// skip/handler callbacks, and sliding or token-bucket algorithms — are not
// implemented here.
package ratelimit

import (
	"strconv"
	"sync"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the rate limiter.
type Options struct {
	// Max is the maximum number of requests permitted per window. Values <= 0
	// default to 60.
	Max int
	// Window is the length of the fixed window. Values <= 0 default to one
	// minute.
	Window time.Duration
	// KeyFunc derives the bucket key for a request. When nil the client IP is
	// used.
	KeyFunc func(req *express.Request) string
	// Now returns the current time. When nil it defaults to time.Now. It is
	// primarily useful for deterministic tests.
	Now func() time.Time
}

type window struct {
	count int
	reset time.Time
}

// New returns rate-limiting middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Max <= 0 {
		o.Max = 60
	}
	if o.Window <= 0 {
		o.Window = time.Minute
	}
	if o.KeyFunc == nil {
		o.KeyFunc = func(req *express.Request) string { return req.IP() }
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
			w = &window{count: 0, reset: t.Add(o.Window)}
			buckets[key] = w
		}
		w.count++
		count := w.count
		reset := w.reset
		mu.Unlock()

		remaining := o.Max - count
		if remaining < 0 {
			remaining = 0
		}
		res.Set("X-RateLimit-Limit", strconv.Itoa(o.Max))
		res.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		res.Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))

		if count > o.Max {
			retry := int(reset.Sub(t).Seconds())
			if retry < 1 {
				retry = 1
			}
			res.Set("Retry-After", strconv.Itoa(retry))
			res.Status(429).Send("Too Many Requests")
			return
		}
		next()
	}
}
