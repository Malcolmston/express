// Package basepath provides middleware that strips a fixed prefix from the
// request path, allowing an application mounted under a sub-path (e.g. behind a
// reverse proxy at "/app") to be written as though it were served from root.
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
