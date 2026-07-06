// Package slowlog provides middleware that logs a warning whenever a request
// takes longer than a configured threshold to complete, helping surface slow
// endpoints. It plays the same role as Node timing loggers such as
// response-time paired with a threshold check, or the "slow request" warnings
// emitted by morgan-style access logs, but distilled to a single concern:
// measure wall-clock handler time and shout only when it crosses a line.
//
// Reach for this middleware as lightweight, always-on performance triage. Mount
// it globally so every route is timed, and the log fills only with the requests
// that actually ran slowly — a cheap way to catch a database call that got slow
// under load, an N+1 query, or an endpoint that degrades in production without
// standing up a full tracing stack. Because the quiet path costs one time.Now
// and one subtraction, it is inexpensive enough to leave enabled everywhere.
//
// Operationally the middleware wraps the rest of the chain and belongs at or
// near the front so its measurement spans the handlers it is meant to observe.
// It records the start time, calls next() to run everything downstream, and on
// return computes the elapsed duration with time.Since. It reads no request or
// response headers and writes none; it does not alter the response or
// short-circuit anything, so it is transparent to clients. When elapsed exceeds
// Options.Threshold it emits one line to Options.Logger of the form "slowlog:
// WARNING <method> <path> took <elapsed> (threshold <t>)", using req.Method and
// req.Path and rounding the elapsed time to the nearest microsecond.
//
// Semantics and defaults are minimal. Threshold defaults to one second when left
// zero or negative, and the comparison is strictly greater-than, so a request
// exactly at the threshold is not logged. Logger defaults to a log.Logger
// writing to os.Stderr with the standard date/time flags; pass your own
// log.Logger to redirect the output, silence timestamps, or fan it into a
// structured sink. The timing is measured around next(), so it includes the work
// of every downstream handler and middleware but not the client's network
// transfer time.
//
// Compared with the richer Node ecosystem this port is intentionally spare. It
// logs only a threshold breach rather than timing every request, exposes no
// Server-Timing or X-Response-Time response header, offers no percentile
// aggregation or sampling, and formats a fixed human-readable line rather than
// structured JSON. If you need per-request metrics or headers, layer a dedicated
// timing middleware alongside it; slowlog's single job is to make slow requests
// visible in the log.
package slowlog

import (
	"log"
	"os"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the slow-request logger.
type Options struct {
	// Threshold is the duration above which a request is considered slow
	// (default 1s).
	Threshold time.Duration
	// Logger receives the warning line. When nil a logger writing to os.Stderr
	// is used.
	Logger *log.Logger
}

// New returns middleware that measures the time spent handling each request and
// logs a warning when it exceeds the threshold.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Threshold <= 0 {
		o.Threshold = time.Second
	}
	logger := o.Logger
	if logger == nil {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		start := time.Now()
		next()
		elapsed := time.Since(start)
		if elapsed > o.Threshold {
			logger.Printf("slowlog: WARNING %s %s took %s (threshold %s)",
				req.Method(), req.Path(), elapsed.Round(time.Microsecond), o.Threshold)
		}
	}
}
