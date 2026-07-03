// Package expires provides express middleware that sets an HTTP Expires header
// a fixed duration into the future.
package expires

import (
	"net/http"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the expires middleware.
type Options struct {
	// Duration is added to the current time to produce the Expires value.
	Duration time.Duration
}

// now is overridable in tests.
var now = time.Now

// New returns middleware that sets the Expires response header to the current
// time plus Duration, formatted as an HTTP date (RFC 1123 GMT).
func New(opts ...Options) express.Handler {
	var d time.Duration
	if len(opts) > 0 {
		d = opts[0].Duration
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		t := now().Add(d).UTC()
		res.Set("Expires", t.Format(http.TimeFormat))
		next()
	}
}
