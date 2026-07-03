// Package throttle provides a per-client token-bucket rate limiter for the
// express framework. Each client (keyed by IP by default) has a bucket that
// refills at a fixed rate up to a burst capacity. A request consumes one token;
// when no token is available the request is rejected with 429 Too Many
// Requests.
//
// Refilling is computed from elapsed wall-clock time via an injectable Now
// function, making the behaviour fully deterministic in tests.
package throttle

import (
	"strconv"
	"sync"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the throttle middleware.
type Options struct {
	// Rate is the number of tokens added per second. Values <= 0 default to 1.
	Rate float64
	// Burst is the maximum number of tokens a bucket may hold. Values <= 0
	// default to 1.
	Burst int
	// KeyFunc derives the bucket key. When nil the client IP is used.
	KeyFunc func(req *express.Request) string
	// Now returns the current time. When nil it defaults to time.Now.
	Now func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

// New returns throttle middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Rate <= 0 {
		o.Rate = 1
	}
	if o.Burst <= 0 {
		o.Burst = 1
	}
	if o.KeyFunc == nil {
		o.KeyFunc = func(req *express.Request) string { return req.IP() }
	}
	now := o.Now
	if now == nil {
		now = time.Now
	}

	var mu sync.Mutex
	buckets := make(map[string]*bucket)

	return func(req *express.Request, res *express.Response, next express.Next) {
		key := o.KeyFunc(req)
		t := now()

		mu.Lock()
		b, ok := buckets[key]
		if !ok {
			b = &bucket{tokens: float64(o.Burst), last: t}
			buckets[key] = b
		} else {
			elapsed := t.Sub(b.last).Seconds()
			if elapsed > 0 {
				b.tokens += elapsed * o.Rate
				if b.tokens > float64(o.Burst) {
					b.tokens = float64(o.Burst)
				}
				b.last = t
			}
		}
		allowed := b.tokens >= 1
		if allowed {
			b.tokens--
		}
		tokens := b.tokens
		mu.Unlock()

		if !allowed {
			// Seconds until one token is available.
			retry := (1 - tokens) / o.Rate
			secs := int(retry)
			if float64(secs) < retry {
				secs++
			}
			if secs < 1 {
				secs = 1
			}
			res.Set("Retry-After", strconv.Itoa(secs))
			res.Status(429).Send("Too Many Requests")
			return
		}
		next()
	}
}
