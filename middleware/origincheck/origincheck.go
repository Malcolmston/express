// Package origincheck provides middleware that rejects requests whose Origin
// header is not in a configured allowlist, mitigating cross-site request
// forgery from untrusted origins.
package origincheck

import (
	"net/url"

	"github.com/malcolmston/express"
)

// Options configures the origin-check middleware.
type Options struct {
	// Allow lists the permitted origin hosts (with or without port), such as
	// "example.com" or "example.com:8443". Required.
	Allow []string
	// Optional, when true, permits requests that carry no Origin header.
	Optional bool
}

// New returns middleware that responds with 403 unless the request's Origin
// header host appears in the allowlist.
func New(opts Options) express.Handler {
	allow := make(map[string]struct{}, len(opts.Allow))
	for _, a := range opts.Allow {
		allow[a] = struct{}{}
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		origin := req.Get("Origin")
		if origin == "" {
			if opts.Optional {
				next()
				return
			}
			res.Status(403).Send("Forbidden")
			return
		}
		u, err := url.Parse(origin)
		if err != nil {
			res.Status(403).Send("Forbidden")
			return
		}
		if _, ok := allow[u.Host]; ok {
			next()
			return
		}
		if _, ok := allow[u.Hostname()]; ok {
			next()
			return
		}
		res.Status(403).Send("Forbidden")
	}
}
