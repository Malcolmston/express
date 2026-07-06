// Package trailingslash provides middleware that enforces a consistent
// trailing-slash policy on request paths by redirecting requests that do not
// conform. It is the express framework's Go analogue of Node middleware such
// as express-slash and connect-slashes: a drop-in express.Handler that either
// adds a trailing slash to every path or strips it, issuing an HTTP redirect
// so that each resource is reachable at exactly one canonical URL.
//
// Reach for this middleware when you care about URL canonicalization -- for
// SEO (search engines treat /about and /about/ as distinct URLs and split
// link equity between them), for cache-hit consistency, or simply to keep your
// route table from having to register both forms of every path. Choosing one
// policy and redirecting the other form guarantees clients and crawlers
// converge on a single spelling of each URL.
//
// Operationally the middleware belongs at the very front of the chain, before
// routing, so that non-conforming URLs are redirected before any route
// matcher runs. On each request it inspects req.Raw.URL.Path. The root path
// "/" (and the empty path) is always passed through untouched via next(). For
// any other path it checks for a trailing slash and, depending on the policy,
// computes a target: Enforce appends "/" to a path that lacks one, while Strip
// trims trailing slashes from a path that has one (collapsing to "/" if the
// result would be empty). If the request already conforms to the policy, the
// middleware simply calls next().
//
// When a redirect is required the original query string (req.Raw.URL.RawQuery)
// is re-appended to the target so it survives the round trip, and the response
// is issued via res.Redirect(status, target); next() is not called on the
// redirect path. The redirect status is Options.Status, which defaults to 301
// Moved Permanently when zero -- appropriate for canonicalization -- though
// callers may prefer 308 to force browsers to preserve the method and body of
// non-GET requests. Options.Enforce and Options.Strip are mutually exclusive:
// when both are false the middleware is a no-op that always calls next(), and
// when both are true Enforce takes precedence.
//
// Compared with the Node originals this port is intentionally small. It
// operates purely on the URL path and query and never consults the Host
// header or protocol, it applies one global policy rather than per-route
// exceptions, and it makes no special allowance for paths that look like files
// (for example "/logo.png"), so if you serve static assets without a trailing
// slash you should mount this middleware only on the route subtrees where the
// policy makes sense.
package trailingslash

import (
	"net/http"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the trailing-slash middleware. Enforce and Strip are
// mutually exclusive; when both are false the middleware is a no-op. When both
// are true, Enforce takes precedence.
type Options struct {
	// Enforce redirects paths without a trailing slash to the slashed form.
	Enforce bool

	// Strip redirects paths with a trailing slash to the unslashed form.
	Strip bool

	// Status is the redirect status code. When zero it defaults to 301.
	Status int
}

// New returns middleware implementing the configured trailing-slash policy.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	status := o.Status
	if status == 0 {
		status = http.StatusMovedPermanently
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		path := req.Raw.URL.Path
		if path == "" || path == "/" {
			next()
			return
		}
		hasSlash := strings.HasSuffix(path, "/")

		var target string
		switch {
		case o.Enforce && !hasSlash:
			target = path + "/"
		case !o.Enforce && o.Strip && hasSlash:
			target = strings.TrimRight(path, "/")
			if target == "" {
				target = "/"
			}
		default:
			next()
			return
		}

		if q := req.Raw.URL.RawQuery; q != "" {
			target += "?" + q
		}
		res.Redirect(status, target)
	}
}
