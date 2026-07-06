// Package latencyheader provides middleware that measures how long a request
// takes to process and records the elapsed milliseconds in a response header
// (X-Response-Latency by default). It is the Go analogue of the small
// "response-time" style middleware common in the Node/Express ecosystem
// (for example the response-time npm package used with connect/express),
// exposing the server-side processing time to clients, proxies, and browser
// developer tools without any external dependencies.
//
// Use this middleware when you want lightweight, always-on latency visibility
// for every response: profiling slow endpoints in development, feeding
// synthetic monitoring that scrapes the header, or letting a downstream load
// balancer or CDN observe per-request processing cost. Because the value is
// computed from a monotonic wall-clock delta and written as a plain integer,
// it is cheap enough to leave enabled in production, and the injectable clock
// makes the timing reproducible in tests.
//
// Register it early in the chain — ideally as the first app.Use handler — so
// the measured interval spans as much of the request lifecycle as possible.
// On entry the handler captures a start timestamp via Options.Now, then
// registers an OnBeforeWrite hook on the response and immediately calls next()
// to run the rest of the stack. When any downstream handler commits the
// response (for example through res.Send or res.JSON), the hook fires, computes
// now-start in milliseconds, and writes it into the configured header just
// before the status line and headers are flushed. Nothing about the request is
// read or mutated; only a single response header is added.
//
// The two options are Header (the response header name, defaulting to
// "X-Response-Latency") and Now (the clock, defaulting to time.Now). The
// elapsed duration is truncated to whole milliseconds via
// time.Duration.Milliseconds, so sub-millisecond requests report 0 and there is
// no fractional component. Because the value is emitted from OnBeforeWrite, a
// handler that never writes a response — or one that panics before writing —
// will not produce the header; conversely a handler that writes exactly once
// gets exactly one measurement. The middleware short-circuits nothing and never
// blocks the chain: it always calls next().
//
// There are no security considerations beyond the mild information disclosure of
// advertising server processing time, which some deployments prefer to strip at
// the edge. Compared with the Node original, this port keeps the essential
// behavior (measure the request and stamp the elapsed time into a header) while
// adopting Go idioms: a functional Now clock for deterministic testing instead
// of a mocked Date, integer-millisecond output rather than the configurable
// digits/formatting hooks of the npm module, and a fixed header default of
// X-Response-Latency rather than X-Response-Time.
package latencyheader

import (
	"strconv"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the latency-header middleware.
type Options struct {
	// Header is the response header name. Defaults to "X-Response-Latency".
	Header string
	// Now returns the current time. When nil it defaults to time.Now. Tests may
	// inject a controllable clock for deterministic results.
	Now func() time.Time
}

// New returns latency-header middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Header == "" {
		o.Header = "X-Response-Latency"
	}
	now := o.Now
	if now == nil {
		now = time.Now
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		start := now()
		res.OnBeforeWrite(func() {
			ms := now().Sub(start).Milliseconds()
			res.Set(o.Header, strconv.FormatInt(ms, 10))
		})
		next()
	}
}
