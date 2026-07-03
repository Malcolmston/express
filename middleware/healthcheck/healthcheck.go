// Package healthcheck provides a liveness/readiness endpoint for the express
// framework. When a request targets the configured path, each registered
// checker is executed and the aggregate result is returned as JSON. If every
// checker succeeds the response status is 200; otherwise it is 503. Requests to
// any other path fall through to the next handler.
package healthcheck

import (
	"net/http"
	"sort"

	"github.com/malcolmston/express"
)

// Options configures the healthcheck middleware.
type Options struct {
	// Path is the endpoint served. Defaults to "/healthz".
	Path string
	// Checkers maps a component name to a function returning nil when healthy.
	Checkers map[string]func() error
}

type result struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// New returns healthcheck middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Path == "" {
		o.Path = "/healthz"
	}

	// Copy checkers so later mutation of the caller's map has no effect.
	checkers := make(map[string]func() error, len(o.Checkers))
	names := make([]string, 0, len(o.Checkers))
	for name, fn := range o.Checkers {
		checkers[name] = fn
		names = append(names, name)
	}
	sort.Strings(names)

	return func(req *express.Request, res *express.Response, next express.Next) {
		if req.Path() != o.Path {
			next()
			return
		}
		checks := make(map[string]string, len(checkers))
		healthy := true
		for _, name := range names {
			if err := checkers[name](); err != nil {
				healthy = false
				checks[name] = err.Error()
			} else {
				checks[name] = "ok"
			}
		}
		out := result{Status: "ok", Checks: checks}
		code := http.StatusOK
		if !healthy {
			out.Status = "unavailable"
			code = http.StatusServiceUnavailable
		}
		res.Status(code).JSON(out)
	}
}
