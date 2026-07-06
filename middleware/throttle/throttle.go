// Package throttle provides a per-client token-bucket rate limiter for the
// express framework. It is the Go analogue of Node rate-limiting middleware
// such as express-rate-limit, express-throttle, or the token-bucket limiters
// built on top of the "limiter" npm module, packaged as a drop-in
// express.Handler. Each client (keyed by IP by default) is given its own
// bucket that refills at a fixed Rate up to a Burst capacity; each request
// consumes one token, and when no token is available the request is rejected
// with 429 Too Many Requests plus a Retry-After header.
//
// Reach for this middleware when you want to protect an endpoint from abuse or
// accidental flooding: login and password-reset routes, expensive search or
// report endpoints, public APIs that must enforce a fair-use quota, or any
// handler whose cost you want to spread out over time. The token-bucket model
// allows short bursts (up to Burst requests back-to-back) while still capping
// the sustained request rate at Rate requests per second, which is friendlier
// to legitimate clients than a hard fixed-window counter.
//
// Operationally the middleware belongs near the front of the chain, before the
// handlers whose work you want to protect. On each request it derives a bucket
// key with Options.KeyFunc (defaulting to req.IP()), refills that key's bucket
// based on the wall-clock time elapsed since its last request, and then tries
// to spend one token. When a token is available it is deducted and next() is
// called so the request proceeds normally; the middleware writes nothing to
// the response in the success path. Buckets are held in an in-memory map
// guarded by a sync.Mutex, created lazily on first sighting of a key and
// seeded full at Burst tokens, so a brand-new client always gets its full
// burst allowance immediately.
//
// When the bucket is empty the request is short-circuited: the middleware
// computes the whole number of seconds until one token will have refilled,
// sets Retry-After to that value (never less than 1), and replies with
// res.Status(429).Send("Too Many Requests") without calling next(). Rate and
// Burst both coerce non-positive values to 1, so an empty Options{} yields a
// strict one-request-per-second limiter. Refill is computed from an injectable
// Options.Now clock (defaulting to time.Now), which is what makes the tests
// deterministic; callers can supply the same hook to reason about limits under
// a controlled clock.
//
// Compared with the Node originals this port is deliberately minimal and
// process-local. The bucket map lives in the memory of a single process and is
// never evicted, so it is unsuitable as-is for a multi-instance deployment
// (where you would back the counters with Redis or similar) and a long-lived
// server with an unbounded key space will accumulate buckets over time. It
// emits only Retry-After -- not the X-RateLimit-Limit / X-RateLimit-Remaining
// family some middleware adds -- does not distinguish routes, and offers no
// allow-list or custom rejection handler; the response body is always the
// fixed "Too Many Requests" string.
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
