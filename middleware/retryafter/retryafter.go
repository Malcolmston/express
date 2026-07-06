// Package retryafter provides middleware that automatically attaches a
// Retry-After header to responses whose status indicates the client should back
// off (429 Too Many Requests and 503 Service Unavailable by default). It has no
// single Node counterpart; instead it packages, as a reusable express.Handler,
// the Retry-After hinting that libraries such as express-rate-limit and
// express's own maintenance/503 handlers emit ad hoc. The header tells
// well-behaved clients and proxies how many seconds to wait before retrying,
// per RFC 9110 section 10.2.3.
//
// Reach for this middleware when your app returns throttling or maintenance
// responses and you want a consistent, machine-readable back-off hint without
// repeating res.Set("Retry-After", ...) at every call site. Mount it once, high
// in the chain, and any downstream handler or rate limiter that produces a 429
// or 503 automatically gains the header. For the occasional handler that emits
// such a status directly and wants a bespoke value, the package also exposes the
// Set helper, which writes the header immediately without the middleware.
//
// Operationally the middleware belongs before the handlers and limiters whose
// statuses it should annotate. On each request it registers a single
// OnBeforeWrite hook and immediately calls next(); it never short-circuits and
// never alters the body or status. The hook runs once, just before the response
// headers are committed, and inspects the final status code via res.StatusCode().
// Only when that status is in the configured set and no Retry-After header has
// already been written does it set the header — so timing the check at commit
// captures the true status a downstream handler ultimately chose.
//
// Two options tune the behavior. Options.Seconds is the value written to the
// header and defaults to 1 when zero or negative. Options.Statuses is the set of
// status codes that should receive the header and defaults to 429 and 503 when
// empty. The middleware deliberately does not overwrite an existing Retry-After
// header, so a handler or limiter that already set a more precise value (for
// example a per-client back-off computed from a token bucket) wins. The Set
// helper follows a similar clamp: a negative seconds argument is coerced to 0
// (Set does not apply the 1-second default), and it always writes the header,
// overwriting any prior value.
//
// Compared with hand-rolled Retry-After handling in the Node ecosystem, this
// port is intentionally narrow. It emits only the integer-seconds form of
// Retry-After, never the HTTP-date form; it treats the value as a constant per
// middleware instance rather than deriving it from live limiter state; and it
// keys purely off the response status, with no awareness of routes, clients, or
// rate-limit windows. For dynamic, per-request back-off values, compute the
// number yourself and use Set (or res.Set) from the handler that knows it.
package retryafter

import (
	"strconv"

	"github.com/malcolmston/express"
)

// Options configures the retry-after middleware.
type Options struct {
	// Seconds is the value written to the Retry-After header. Values <= 0
	// default to 1.
	Seconds int
	// Statuses lists the status codes that should receive a Retry-After header.
	// When empty it defaults to 429 and 503.
	Statuses []int
}

// Set writes a Retry-After header with the given number of seconds directly on
// res. It is a convenience helper for handlers that emit a 429/503 themselves.
func Set(res *express.Response, seconds int) {
	if seconds < 0 {
		seconds = 0
	}
	res.Set("Retry-After", strconv.Itoa(seconds))
}

// New returns retry-after middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Seconds <= 0 {
		o.Seconds = 1
	}
	if len(o.Statuses) == 0 {
		o.Statuses = []int{429, 503}
	}
	match := make(map[int]bool, len(o.Statuses))
	for _, s := range o.Statuses {
		match[s] = true
	}
	value := strconv.Itoa(o.Seconds)

	return func(req *express.Request, res *express.Response, next express.Next) {
		res.OnBeforeWrite(func() {
			if match[res.StatusCode()] && res.GetHeader("Retry-After") == "" {
				res.Set("Retry-After", value)
			}
		})
		next()
	}
}
