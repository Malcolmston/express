// Package version provides middleware that exposes an application's version.
// It sets an X-Version header on every response and, when a request targets the
// configured path, replies with a small JSON document reporting the version.
// There is no single canonical Node package this ports; it distils a common
// operational pattern seen across Express deployments — advertising the running
// build both as a response header and via a dedicated introspection endpoint
// (comparable to hand-rolled "/version" routes and helpers like
// "express-version-route") — into one small, dependency-free middleware.
//
// Use it to make deploys observable: a load balancer, uptime probe, or curl
// against the version path can confirm which build is live, and every ordinary
// response carries the version header so clients and logs can correlate
// behaviour with a release. It pairs naturally with a healthcheck endpoint.
// Because the version string is emitted to all callers, treat it as public
// information and avoid encoding anything sensitive in it.
//
// Mount it early with app.Use so the header is applied before your routes run.
// On every request the middleware first sets the configured header to the
// version via res.Set. It then compares req.Path() against the configured
// path; on an exact match it short-circuits the chain by replying
// res.Status(200).JSON(payload) — a JSON object of the shape
// {"version":"..."} — and returns without calling next. For any other path it
// calls next() so the request continues, meaning the header decorates
// responses produced by downstream handlers as well.
//
// All three fields of Options have defaults, so New() with no arguments is
// valid: Version defaults to "unknown", Path defaults to "/version", and
// Header defaults to "X-Version". Supplying an Options overrides any non-empty
// field while leaving the rest at their defaults. Note that the path match is
// an exact, case-sensitive equality against req.Path() with no trailing-slash
// normalization or method filtering, so the JSON endpoint answers any HTTP
// method on exactly that path.
//
// Compared with ad-hoc Node equivalents this port is deliberately compact: the
// JSON payload contains only the version field (no name, commit, build time,
// or uptime), the response is not content-negotiated, and there is no
// per-request version selection or route-parameter matching. It is a
// fixed-string reporter configured once at construction. If you need a richer
// document, add your own handler and read the same version string.
package version

import "github.com/malcolmston/express"

// Options configures the version middleware.
type Options struct {
	// Version is the version string reported. Defaults to "unknown".
	Version string
	// Path is the endpoint that returns the version as JSON. Defaults to
	// "/version".
	Path string
	// Header is the response header name. Defaults to "X-Version".
	Header string
}

type payload struct {
	Version string `json:"version"`
}

// New returns version middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Version == "" {
		o.Version = "unknown"
	}
	if o.Path == "" {
		o.Path = "/version"
	}
	if o.Header == "" {
		o.Header = "X-Version"
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set(o.Header, o.Version)
		if o.Path != "" && req.Path() == o.Path {
			res.Status(200).JSON(payload{Version: o.Version})
			return
		}
		next()
	}
}
