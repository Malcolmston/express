// Package readiness provides a readiness-probe endpoint for the express
// framework. Requests to the configured path return 200 when the application
// reports itself ready and 503 otherwise. Requests to any other path fall
// through to the next handler.
package readiness

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the readiness middleware.
type Options struct {
	// Ready reports whether the application is ready to serve traffic. When nil
	// the application is always considered ready.
	Ready func() bool
	// Path is the endpoint served. Defaults to "/readyz".
	Path string
}

// New returns readiness middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Path == "" {
		o.Path = "/readyz"
	}
	ready := o.Ready
	if ready == nil {
		ready = func() bool { return true }
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if req.Path() != o.Path {
			next()
			return
		}
		if ready() {
			res.Status(http.StatusOK).Send("ready")
			return
		}
		res.Status(http.StatusServiceUnavailable).Send("not ready")
	}
}
