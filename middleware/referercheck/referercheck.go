// Package referercheck provides middleware that rejects requests whose Referer
// header host is not in a configured allowlist.
package referercheck

import (
	"net/url"

	"github.com/malcolmston/express"
)

// Options configures the referer-check middleware.
type Options struct {
	// Allow lists the permitted referer hosts (with or without port).
	// Required.
	Allow []string
	// Optional, when true, permits requests that carry no Referer header.
	Optional bool
}

// New returns middleware that responds with 403 unless the request's Referer
// header host appears in the allowlist. When Optional is set, requests without
// a Referer header are allowed through.
func New(opts Options) express.Handler {
	allow := make(map[string]struct{}, len(opts.Allow))
	for _, a := range opts.Allow {
		allow[a] = struct{}{}
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		ref := req.Get("Referer")
		if ref == "" {
			if opts.Optional {
				next()
				return
			}
			res.Status(403).Send("Forbidden")
			return
		}
		u, err := url.Parse(ref)
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
