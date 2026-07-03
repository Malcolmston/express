// Package uptime provides middleware that reports how long the middleware (and,
// by extension, the process) has been running. It sets an X-Uptime header in
// whole seconds on every response and exposes a Since accessor returning the
// elapsed duration.
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
