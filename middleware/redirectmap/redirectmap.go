// Package redirectmap provides middleware that redirects requests whose path
// matches an entry in a static lookup table. It is handy for migrating old
// URLs to new locations without adding a route for each one.
package redirectmap

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the redirect-map middleware.
type Options struct {
	// Map associates request paths with destination URLs. A request whose
	// path is a key is redirected to the corresponding value.
	Map map[string]string

	// Status is the HTTP status code used for redirects. When zero it
	// defaults to 302 (Found).
	Status int
}

// New returns middleware that redirects any request whose path is present in
// the map to the mapped destination; other requests fall through to next.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	status := o.Status
	if status == 0 {
		status = http.StatusFound
	}
	// Copy the map so later mutation by the caller cannot affect behavior.
	m := make(map[string]string, len(o.Map))
	for k, v := range o.Map {
		m[k] = v
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if dest, ok := m[req.Path()]; ok {
			res.Redirect(status, dest)
			return
		}
		next()
	}
}
