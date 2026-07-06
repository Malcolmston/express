// Package basepath provides middleware that strips a fixed prefix from the
// request path, allowing an application mounted under a sub-path (for example
// behind a reverse proxy at "/app") to be written as though it were served
// from root. It is the Go analogue of Node helpers such as express-basepath
// and the mounting behavior of Express sub-apps, packaged as a drop-in
// express.Handler that rewrites the router's match path before downstream
// routes run.
//
// Use this middleware when the public URL space and the application's internal
// route definitions differ by a constant prefix. A common case is a service
// deployed under "/app" by an upstream proxy or gateway while its handlers are
// authored against clean paths like "/users" and "/health". Instead of
// prefixing every route registration, mount basepath first and let each
// request arrive at the routes with the prefix already removed, keeping route
// definitions portable across mount points.
//
// Operationally the middleware belongs at the very front of the chain, before
// any routing. On each request it inspects req.Raw.URL.Path: when the path
// equals the prefix or begins with prefix + "/", it computes the remainder,
// substitutes "/" when the remainder would be empty, and calls
// req.SetPath(stripped) so the router matches downstream routes against the
// shortened path. It then calls next() to continue the chain. The prefix is
// normalized once when the handler is built — a missing leading slash is added
// and any trailing slash is trimmed — so "app", "/app", and "/app/" all behave
// identically.
//
// The two options are Options.Prefix and Options.Strict. An empty Prefix
// normalizes to a no-op that leaves every path unchanged. Requests whose path
// does not fall under the prefix are handled according to Strict: when false
// (the default) they pass through untouched via next(), which is useful when
// other routes or middleware live outside the mounted sub-path; when true they
// are rejected with 404 via res.SendStatus, enforcing that only the sub-path is
// served. Note that only the router match path is rewritten through SetPath;
// req.Raw.URL.Path itself is not mutated, so middleware reading the raw URL
// still observes the original, un-stripped path.
//
// Compared with the Node originals, this port focuses solely on inbound path
// rewriting for routing: it does not rewrite outbound redirect Location
// headers, adjust generated links, or set a base-URL value for templates.
// Applications that need those behaviors should compose additional middleware,
// but for the core task of matching clean routes behind a fixed prefix this
// port matches the expected mount-and-strip semantics.
package basepath

import (
	"net/http"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the basepath middleware.
type Options struct {
	// Prefix is the path prefix to strip, e.g. "/app". A missing leading slash
	// is added automatically and a trailing slash is ignored.
	Prefix string

	// Strict, when true, responds 404 to requests that do not begin with
	// Prefix. When false such requests are passed through unchanged.
	Strict bool
}

// New returns middleware that strips Options.Prefix from req.Raw.URL.Path.
func New(opts Options) express.Handler {
	prefix := opts.Prefix
	if prefix == "" {
		prefix = "/"
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = strings.TrimRight(prefix, "/")

	return func(req *express.Request, res *express.Response, next express.Next) {
		path := req.Raw.URL.Path
		if prefix == "" || path == prefix || strings.HasPrefix(path, prefix+"/") {
			stripped := strings.TrimPrefix(path, prefix)
			if stripped == "" {
				stripped = "/"
			}
			// SetPath updates the router's match path so downstream routes are
			// matched against the stripped path.
			req.SetPath(stripped)
			next()
			return
		}
		if opts.Strict {
			res.SendStatus(http.StatusNotFound)
			return
		}
		next()
	}
}
