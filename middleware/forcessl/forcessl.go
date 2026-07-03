// Package forcessl provides middleware that redirects insecure HTTP requests
// to their HTTPS equivalent, preserving the host, path, and query string.
package forcessl

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the force-SSL middleware.
type Options struct {
	// Enabled turns the redirect on. Because the zero value of a bool is
	// false, New defaults Enabled to true when Options are omitted entirely;
	// pass Options{Enabled: false} to explicitly disable.
	Enabled bool
}

// New returns middleware that redirects http requests to https with a 301
// (Moved Permanently). Secure requests, and all requests when disabled, pass
// through untouched. The request is considered secure when it arrived over TLS
// or carries X-Forwarded-Proto: https (both handled by req.Secure).
func New(opts ...Options) express.Handler {
	o := Options{Enabled: true}
	if len(opts) > 0 {
		o = opts[0]
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if !o.Enabled || req.Secure() {
			next()
			return
		}
		res.Redirect(http.StatusMovedPermanently, "https://"+req.Raw.Host+req.OriginalURL())
	}
}
