// Package latencyheader provides middleware that measures how long a request
// takes to process and records the elapsed milliseconds in a response header
// (X-Response-Latency by default). The header is set through an OnBeforeWrite
// hook so it is emitted just before the response is committed.
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
