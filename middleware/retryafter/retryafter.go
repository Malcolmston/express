// Package retryafter provides middleware that automatically attaches a
// Retry-After header to responses whose status indicates the client should
// back off (429 Too Many Requests and 503 Service Unavailable). The header is
// applied via an OnBeforeWrite hook so that it reflects the final status code
// chosen by downstream handlers.
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
