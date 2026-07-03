// Package healthz provides a minimal liveness endpoint. Requests to the
// configured path receive a plain 200 "ok" response; all other requests fall
// through to the next handler.
package healthz

import "github.com/malcolmston/express"

// Options configures the healthz middleware.
type Options struct {
	// Path is the endpoint served. Defaults to "/healthz".
	Path string
	// Body is the response body. Defaults to "ok".
	Body string
}

// New returns healthz middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Path == "" {
		o.Path = "/healthz"
	}
	if o.Body == "" {
		o.Body = "ok"
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		if req.Path() == o.Path {
			res.Status(200).Send(o.Body)
			return
		}
		next()
	}
}
