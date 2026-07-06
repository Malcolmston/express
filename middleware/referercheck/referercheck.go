// Package referercheck provides middleware that rejects requests whose Referer
// header host is not in a configured allowlist. It ports the classic Node
// referer-based CSRF guard — the pattern behind packages and hand-written
// middleware that compare req.headers.referer against a set of permitted hosts
// and 403 anything else — into a single express handler backed by a host set.
//
// Use it as a lightweight, defense-in-depth guard on state-changing endpoints
// (form posts, mutating APIs) to ensure the request was initiated from a page
// your own site served. Because a browser sets the Referer for same-site
// navigations, requiring a known host raises the bar for cross-site request
// forgery and hotlinking without the token plumbing of a full CSRF library. It
// is best treated as one layer among several rather than a sole defense, since
// the Referer header can be suppressed or stripped by proxies and privacy tools.
//
// Mechanically the handler reads the "Referer" header on each request. A missing
// header is rejected with 403 "Forbidden" unless Optional is set, in which case
// it calls next() and lets the request through. A present header is parsed with
// net/url; a parse error is a 403. On a successful parse it checks the allowlist
// against both u.Host (host with any port) and u.Hostname() (bare host), calling
// next() on a match and responding 403 otherwise. On rejection it writes the
// response itself and does not call next(), short-circuiting the chain, so it
// should be registered via app.Use ahead of the routes it protects.
//
// The Allow option is required and holds the permitted hosts; entries are
// matched by exact equality after parsing, so an allowlist of "example.com"
// admits both a bare "example.com" referer and one carrying a port only when the
// stored entry itself matches — list the port-qualified form (for example
// "example.com:8443") explicitly if you need it. Optional defaults to false,
// meaning absent Referer headers are denied; set it to true to tolerate clients
// and privacy configurations that omit the header, trading strictness for
// compatibility, as the MissingRefererOptional test shows. There is no scheme,
// path, or subdomain wildcarding — matching is host-exact by design.
//
// Parity with the Node original is behavioral: like the typical referer CSRF
// middleware it allowlists hosts, returns 403 on mismatch, malformed, or (by
// default) missing Referer, and offers an opt-in to permit the missing case.
// It intentionally omits richer policies some libraries add — regex or
// subdomain matching, per-route configuration, or combining Referer with Origin
// header checks — keeping the contract a small, auditable host comparison that
// callers can compose with the sibling referer package for capture.
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
