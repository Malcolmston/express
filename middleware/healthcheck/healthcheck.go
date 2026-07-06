// Package healthcheck provides a liveness/readiness endpoint for the express
// framework. It is a stdlib-only analogue of Node health-check middlewares such
// as "express-healthcheck" and the "@cloudnative/health-connect" / Terminus
// style probes: a single mounted handler owns one path, runs a set of
// user-supplied component checks, and reports their aggregate result as JSON
// suitable for a load balancer, Kubernetes readiness probe, or uptime monitor.
//
// Use this rather than the sibling healthz package when a bare "the process is
// up" signal is not enough and you need to verify that dependencies — a
// database, a cache, a downstream API — are actually reachable before declaring
// the instance ready to receive traffic. Each dependency is expressed as a
// named function returning error; returning nil means healthy and returning a
// non-nil error means unhealthy, with the error text surfaced in the response so
// operators can see *why* a check failed.
//
// The middleware is typically mounted globally with app.Use, at or near the top
// of the chain. On every request it compares req.Path() against the configured
// Path. If they differ it calls next() and the request falls through to the rest
// of the application untouched — so it is safe to mount alongside your normal
// routes. If they match it short-circuits the chain (it never calls next()),
// runs every checker, and writes the response with res.Status(code).JSON(out).
// The JSON body has the shape {"status": "...", "checks": {name: detail}} where
// each check's value is "ok" on success or the error string on failure. When all
// checkers pass the top-level status is "ok" and the HTTP code is 200; if any
// checker fails the status becomes "unavailable" and the code is 503
// (Service Unavailable).
//
// Several semantics are worth noting. The Path defaults to "/healthz" when left
// empty. The checker map is copied at construction time, so mutating the map you
// passed in after calling New has no effect on the running endpoint; this also
// means checks are a fixed set decided at startup. Checker names are sorted so
// the set of checks run is stable and order-independent, though JSON object key
// order in the emitted body is not itself significant. Every checker runs on
// every probe (there is no caching, timeout, or parallelism), so keep each check
// fast and non-blocking, and guard against a slow dependency stalling the probe —
// a checker that hangs will hang the health endpoint. Checkers are invoked
// synchronously in the request goroutine and must be safe for concurrent calls
// across simultaneous probes.
//
// Compared with the Node originals this port keeps the core "run named checks,
// aggregate to 200/503, emit JSON" contract but omits extras those packages
// sometimes offer, such as per-check timeouts, response caching, custom body
// templates, uptime/version metadata, or separate liveness and readiness routes.
// If you need only a static up/down signal with no dependency checks, prefer the
// lighter healthz package.
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
