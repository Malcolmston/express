// Package hostcheck provides middleware that guards against Host header
// attacks by permitting only requests whose hostname is in a configured
// allowlist.
package hostcheck

import "github.com/malcolmston/express"

// Options configures the host-check middleware.
type Options struct {
	// Allow lists the permitted hostnames (without port). Required.
	Allow []string
	// Status is the response code used when a host is not allowed. Defaults
	// to 421 (Misdirected Request); 400 is a common alternative.
	Status int
}

// New returns middleware that rejects requests whose Hostname is not present in
// the allowlist, defending against Host header spoofing.
func New(opts Options) express.Handler {
	status := opts.Status
	if status == 0 {
		status = 421
	}
	allow := make(map[string]struct{}, len(opts.Allow))
	for _, h := range opts.Allow {
		allow[h] = struct{}{}
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		if _, ok := allow[req.Hostname()]; ok {
			next()
			return
		}
		res.Status(status).Send("Misdirected Request")
	}
}
