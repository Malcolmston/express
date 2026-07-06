// Package concurrencylimit provides express middleware that caps the number of
// requests processed concurrently and sheds load once that cap is reached. It
// is the Go analogue of the Node load-shedding pattern used by middleware such
// as express-queue (run with a zero-length queue) or the connection-limiting
// behaviour of toobusy-js, packaged as a drop-in express.Handler. Rather than
// smoothing traffic over time like a rate limiter, it bounds the instantaneous
// in-flight request count so a burst cannot overwhelm downstream resources.
//
// Use this middleware to protect a service whose expensive work (database
// pools, external APIs, CPU-bound handlers) has a known safe concurrency
// ceiling. When more requests arrive than the ceiling allows, failing them
// fast with a 503 is often healthier than letting them queue unboundedly and
// drive latency and memory upward. Mount it globally with app.Use to guard the
// whole application, or attach it to a specific router or path prefix to fence
// off only the costly endpoints while leaving cheap ones unthrottled.
//
// Operationally the middleware sits at the front of the chain and is backed by
// a buffered channel of capacity Max that acts as a counting semaphore. On
// each request it performs a non-blocking send into the channel: if a slot is
// free the token is taken, next() is invoked to run the rest of the chain, and
// a deferred receive returns the token when the handler completes — including
// when a downstream handler panics, because the release runs in a defer.
// Requests are never parked waiting for a slot; the semaphore is only ever
// probed, never blocked on.
//
// When no slot is available the select falls through to its default case and
// the request is short-circuited immediately: the middleware writes a 503
// Service Unavailable response carrying Options.Message and never calls next(),
// so no downstream handler runs. Options.Max defaults to 1 for any value less
// than or equal to zero, meaning an unconfigured limiter serializes requests
// one at a time. Options.Message defaults to "Service Unavailable". The
// semaphore is created once when New is called and shared across every request
// the returned handler serves, so a single New instance enforces one global
// limit; construct separate limiters for independent pools.
//
// Compared with the Node originals, this port keeps the same reject-when-full,
// fail-fast semantics but is intentionally minimal. It does not queue or delay
// overflow requests, does not measure event-loop lag or system load the way
// toobusy-js does, and does not emit Retry-After or per-client accounting; the
// only knobs are the concurrency ceiling and the rejection body.
package concurrencylimit

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the concurrency limiter.
type Options struct {
	// Max is the maximum number of in-flight requests. Values <= 0 default to 1.
	Max int
	// Message is the body sent when the limit is exceeded. Defaults to
	// "Service Unavailable".
	Message string
}

// New returns concurrency-limiting middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Max <= 0 {
		o.Max = 1
	}
	if o.Message == "" {
		o.Message = "Service Unavailable"
	}

	sem := make(chan struct{}, o.Max)

	return func(req *express.Request, res *express.Response, next express.Next) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			next()
		default:
			res.Status(http.StatusServiceUnavailable).Send(o.Message)
		}
	}
}
