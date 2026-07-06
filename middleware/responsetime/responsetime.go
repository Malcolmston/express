// Package responsetime provides express middleware that measures how long a
// request takes to process and reports it in an X-Response-Time header. It is
// the express framework's Go port of the popular Connect/Express response-time
// middleware (the npm "response-time" package), reproducing its default
// behavior of stamping each response with the elapsed handling time in
// milliseconds.
//
// Reach for this middleware to get cheap, always-on latency visibility during
// development and in production: the header lets you spot slow endpoints from a
// browser's network panel, a curl -I, or a log of response headers without
// wiring up a metrics backend. It is a diagnostic aid rather than a full
// observability solution — for percentiles, histograms, or per-route timing you
// would still export to a real metrics system — but for a quick read on how long
// a request spent inside the app it is effectively free.
//
// Operationally the middleware belongs as early as possible in the chain, ideally
// the very first Use, so the measured window spans as much of the request
// lifecycle as possible. On each request it records a start timestamp, registers
// an OnBeforeWrite hook, and immediately calls next() so downstream handlers run
// normally. The hook fires once, at the moment the response headers are about to
// be committed: it computes the elapsed time, converts it to milliseconds, and
// sets the X-Response-Time header (whose name is exported as HeaderName) to a
// value like "12.34ms" — a fixed two-decimal float with an "ms" suffix.
//
// Timing the header via OnBeforeWrite rather than after next() returns is
// deliberate: it ensures the duration is captured before the client sees the
// response and that the header is added while headers are still mutable, since a
// header set after the body has begun streaming would be ignored. The window
// therefore covers work up to the header commit; body streaming that happens
// after headers are flushed is not included. The middleware only ever writes the
// single X-Response-Time header and never short-circuits — it always calls
// next() and defers entirely to the rest of the chain to produce the response.
//
// Compared with the Node original this port keeps the same header name, the same
// millisecond unit, and the same two-digit precision, but is fixed rather than
// configurable: there is no option to change the header name, the number of
// decimal digits, the time unit, or to supply a custom formatting callback, and
// it does not suffix the value with a Server-Timing entry. The internal clock is
// swappable only within the package (an unexported now variable) to make the
// timing deterministic in tests; callers cannot override it.
package responsetime

import (
	"strconv"
	"time"

	"github.com/malcolmston/express"
)

// HeaderName is the response header carrying the measured duration.
const HeaderName = "X-Response-Time"

// now is overridable in tests.
var now = time.Now

// New returns middleware that records the time spent handling a request and
// sets X-Response-Time (in milliseconds, e.g. "12.34ms"). The header is
// written via OnBeforeWrite so the timing reflects work up to the moment the
// response headers are committed.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		start := now()
		res.OnBeforeWrite(func() {
			elapsed := now().Sub(start)
			ms := float64(elapsed) / float64(time.Millisecond)
			res.Set(HeaderName, strconv.FormatFloat(ms, 'f', 2, 64)+"ms")
		})
		next()
	}
}
