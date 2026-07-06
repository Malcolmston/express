// Package vhost provides virtual-host middleware that dispatches requests to a
// dedicated handler based on the request's hostname. It supports exact matches
// and a leading "*." wildcard that matches any single-or-multi-label subdomain.
// It is a port of the Node "vhost" connect/Express middleware: a way to run
// several logical sites — an API host, a docs host, per-tenant subdomains —
// behind a single server and route each incoming request to the right handler
// purely by its Host header.
//
// Reach for it when one process must serve multiple domains or subdomains and
// you want a clean split between their handler trees rather than branching on
// the hostname inside every route. Typical uses are directing
// "api.example.com" to an API sub-application, giving "*.tenant.example.com"
// a shared multi-tenant handler, or carving "admin.example.com" off from the
// public site. Because matching is driven by the client-supplied Host header,
// vhost is a routing convenience, not an authentication or trust boundary.
//
// Mount it with app.Use, usually near the top of the chain so host dispatch
// happens before generic routes. On each request it reads the hostname via
// req.Hostname() (already stripped of any port), lower-cases it, and tests it
// against the configured Host. On a match it invokes Options.Handler with the
// same req, res and next, delegating the request to that handler; on no match
// — or if Handler is nil — it calls next() so the request falls through to the
// rest of the chain. It writes no headers and produces no body of its own; the
// dispatched handler decides how to respond and whether to continue the chain.
//
// Matching is case-insensitive. An exact Host such as "api.example.com"
// matches only that name. A leading "*." makes Host a wildcard: the internal
// implementation strips the "*" to a "." suffix, so "*.example.com" matches
// "a.example.com" and "a.b.example.com" (any depth of subdomain) but pointedly
// not the bare "example.com", because a real label must precede the dot. There
// is no partial-label or mid-string wildcard support, and the handler is only
// invoked when it is non-nil, so a matching Host with a nil Handler simply
// falls through.
//
// Compared with the Node "vhost" package this port keeps the core exact and
// "*." wildcard dispatch but is narrower in a few respects: Options.Host takes
// a plain hostname pattern rather than an arbitrary regular expression, there
// is no exposure of a captured wildcard segment to the handler (the Node
// version populates req.vhost), and a single New call registers exactly one
// host-to-handler mapping — compose several with successive app.Use calls to
// cover multiple hosts.
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
