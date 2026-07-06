// Package uptime provides middleware that reports how long the middleware --
// and, by extension, the process serving requests -- has been running. It has
// no single Node original but plays the role of the small "uptime" helpers and
// health-endpoint fields common in Express apps (for example the process.uptime()
// value surfaced by express-status-monitor or a custom /healthz handler),
// packaged here as a drop-in express.Handler paired with a programmatic
// accessor.
//
// Reach for this middleware when you want a cheap, always-available liveness
// signal: load balancers and orchestrators can read the header to gauge how
// long an instance has been up, operators can spot a process that recently
// restarted (a crash-loop shows as a header that keeps resetting to a small
// number), and dashboards can chart per-instance age. Because it adds a single
// header and never blocks, it is safe to mount globally in front of the whole
// application.
//
// Operationally New records a start timestamp at construction time and returns
// two values: the express.Handler to mount and a since accessor of type
// func() time.Duration. Mount the handler early in the chain so the header is
// present on as many responses as possible. On each request the handler
// computes the elapsed time since start, truncates it to whole seconds, writes
// it to the response header named by Options.Header via res.Set, and then
// always calls next(); it never short-circuits, inspects the request, or
// alters status or body, so it composes freely with any other middleware.
//
// The accessor returned alongside the handler reports the same elapsed
// duration at full time.Duration precision, letting application code read the
// uptime directly -- for instance to include it in a JSON /status payload --
// without parsing the header back out. Both the header and the accessor are
// driven by the same injectable Options.Now clock (defaulting to time.Now),
// which is what makes the behavior testable under a controllable clock and
// keeps the header and accessor perfectly consistent with each other.
//
// Two option fields tune the behavior: Header (the response header name,
// defaulting to "X-Uptime") and Now (the clock). Note that "uptime" here means
// time since this middleware instance was created, not the true operating-
// system process start time; if you construct the middleware well after
// program start the two will differ, so build it as early as possible if you
// need them to coincide. The port is intentionally minimal: it emits seconds
// as a plain integer string with no ISO-8601 or human-readable formatting, and
// exposes no built-in HTTP endpoint -- wire the accessor into your own route
// if you want to serve the value as a body.
package uptime

import (
	"strconv"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the uptime middleware.
type Options struct {
	// Header is the response header name. Defaults to "X-Uptime".
	Header string
	// Now returns the current time. When nil it defaults to time.Now. Tests may
	// inject a controllable clock.
	Now func() time.Time
}

// New returns uptime middleware together with a Since accessor reporting the
// elapsed time since the middleware was created.
func New(opts ...Options) (express.Handler, func() time.Duration) {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Header == "" {
		o.Header = "X-Uptime"
	}
	now := o.Now
	if now == nil {
		now = time.Now
	}
	start := now()

	since := func() time.Duration { return now().Sub(start) }

	handler := func(req *express.Request, res *express.Response, next express.Next) {
		secs := int64(since() / time.Second)
		res.Set(o.Header, strconv.FormatInt(secs, 10))
		next()
	}
	return handler, since
}
