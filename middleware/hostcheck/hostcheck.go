// Package hostcheck provides middleware that guards against Host header
// attacks by permitting only requests whose hostname is in a configured
// allowlist. It plays the role of the "allowed hosts" / host-validation
// middleware common in the Node ecosystem (comparable to Django's
// ALLOWED_HOSTS or the host check in webpack-dev-server), adapted to this
// Express-style framework with only the standard library.
//
// Use it whenever your application trusts the incoming Host header for
// anything — generating absolute URLs, password-reset links, cache keys, or
// virtual-host routing. An attacker who can set an arbitrary Host header can
// otherwise poison those values (Host header injection / web cache poisoning)
// even when the request reaches your app through a proxy. Restricting requests
// to a fixed set of expected hostnames removes that class of attack.
//
// In the chain the middleware should be registered very early, before any
// handler that reads the host or emits absolute links, so rejected requests
// never reach application logic. On each request it reads the request hostname
// via req.Hostname(), which returns the Host header with any :port stripped,
// and looks it up in an allowlist set built once at construction. If the host
// is allowed it calls next() and the request proceeds normally; otherwise it
// short-circuits by writing the configured status and a "Misdirected Request"
// body and does not call next().
//
// Behavior is configured through Options. Allow is the required slice of exact
// hostnames to permit; matching is case-sensitive and port-insensitive (the
// port is removed before comparison), and there is no wildcard or suffix
// matching, so list every host explicitly. Status sets the rejection code and
// defaults to 421 (Misdirected Request), the semantically correct response for
// a host the server is not configured to serve; 400 (Bad Request) is a common
// alternative. Note that an empty Allow list rejects every request, so the
// middleware fails closed.
//
// Parity note: unlike some Node host-check libraries this port does not build
// its allowlist from environment variables, support wildcard patterns, or
// distinguish tunnel/loopback hosts. It performs a single exact-match lookup
// against the explicit Allow slice and nothing more, keeping the security model
// simple and auditable.
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
