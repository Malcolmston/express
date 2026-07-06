// Package healthz provides a minimal liveness endpoint. It is a stdlib-only
// analogue of the tiny Node "express-healthcheck" style middleware: a single
// handler owns one path and answers it with a fixed plain-text 200 response,
// signaling only that the process is up and able to serve requests. It performs
// no dependency checks — for that, use the sibling healthcheck package.
//
// Use healthz for the common "is this instance alive?" probe wired to a load
// balancer, a container orchestrator's liveness check, or an uptime monitor.
// Because it returns immediately with a static body and never touches external
// systems, it is cheap and cannot itself be made to fail by a slow dependency,
// which is exactly what you want from a liveness (as opposed to readiness)
// probe. Reach for healthcheck instead when the endpoint must verify that
// downstream dependencies are reachable before reporting healthy.
//
// The middleware is normally mounted globally with app.Use, at or near the top
// of the chain. On each request it compares req.Path() against the configured
// Path. On a match it short-circuits the chain — it writes res.Status(200).Send
// with the configured body and does not call next(). On any other path it calls
// next() and the request falls through to the rest of the application, so it is
// safe to mount alongside your normal routes. It reads no request state, sets no
// custom headers beyond what res.Send implies, and inspects no method, so it
// answers the health path for every HTTP verb.
//
// Both options have defaults: Path defaults to "/healthz" and Body defaults to
// "ok" when left empty, so New() with no arguments serves "ok" at "/healthz".
// The status code is always 200 and is not configurable. Path matching is an
// exact string comparison against req.Path(), so trailing slashes and case must
// match precisely and no prefix or pattern matching is performed.
//
// Compared with the Node originals this port keeps only the essential
// up/down-with-fixed-body behavior and omits extras such as custom status codes,
// per-request handler callbacks, JSON bodies, or method filtering. It is
// intentionally the smallest possible health endpoint; when you outgrow it,
// graduate to the healthcheck package, which adds named dependency checks and a
// 200/503 JSON contract.
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
