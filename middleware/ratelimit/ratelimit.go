// Package ratelimit provides a fixed-window, per-client rate limiter for the
// express framework. Each client (keyed by IP by default) is allowed a maximum
// number of requests within a rolling fixed window; requests beyond the limit
// receive a 429 Too Many Requests response carrying standard rate-limit
// headers.
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
