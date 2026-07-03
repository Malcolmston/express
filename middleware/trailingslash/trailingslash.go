// Package trailingslash provides middleware that enforces a consistent
// trailing-slash policy by redirecting requests that do not conform. It can
// either add a trailing slash to every path or strip it, preserving the query
// string. The root path "/" is always left untouched.
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
