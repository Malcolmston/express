// Package vhost provides virtual-host middleware that dispatches requests to a
// dedicated handler based on the request's hostname. It supports exact matches
// and a leading "*." wildcard that matches any single-or-multi-label subdomain.
package vhost

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the vhost middleware.
type Options struct {
	// Host is the hostname to match, e.g. "api.example.com". A leading "*."
	// makes it a wildcard: "*.example.com" matches "a.example.com" and
	// "a.b.example.com" but not the bare "example.com".
	Host string

	// Handler runs when the request's hostname matches Host.
	Handler express.Handler
}

// New returns middleware that invokes Options.Handler for requests whose
// hostname matches Options.Host; non-matching requests fall through to next.
func New(opts Options) express.Handler {
	host := strings.ToLower(opts.Host)
	wildcard := strings.HasPrefix(host, "*.")
	suffix := ""
	if wildcard {
		// ".example.com" — the dot ensures we match a real subdomain.
		suffix = host[1:]
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		name := strings.ToLower(req.Hostname())
		match := false
		if wildcard {
			match = strings.HasSuffix(name, suffix) && len(name) > len(suffix)
		} else {
			match = name == host
		}
		if match && opts.Handler != nil {
			opts.Handler(req, res, next)
			return
		}
		next()
	}
}
