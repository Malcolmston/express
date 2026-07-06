// Package origincheck provides middleware that rejects requests whose Origin
// header is not in a configured allowlist, mitigating cross-site request
// forgery (CSRF) and other cross-origin abuse from untrusted callers. It
// implements the same defensive pattern as Node origin-verification middleware
// such as the Origin/Referer checks in csurf-style and koa-csrf-style guards:
// rather than issuing and validating tokens, it trusts the browser-set Origin
// header to decide whether a request may proceed.
//
// Use it on state-changing endpoints (or an entire app) that are consumed only
// by a known set of first-party front ends. Because browsers attach the Origin
// header automatically on cross-origin and non-GET requests and forbid pages
// from forging it, an allowlist of expected origins is a cheap, stateless way
// to block requests initiated by an attacker's site. It is complementary to,
// not a replacement for, token-based CSRF protection and standard
// authentication.
//
// Mechanically the middleware runs before your route handlers, typically via
// app.Use. For each request it reads the Origin header with req.Get("Origin"),
// parses it as a URL, and checks the result against a set built once from
// Options.Allow. A request is allowed to continue — the middleware calls next —
// when the parsed origin's host (u.Host, which includes any port) or its
// hostname (u.Hostname(), port stripped) is present in the allowlist. Any other
// outcome short-circuits the chain with res.Status(403).Send("Forbidden") and
// next is never called, so downstream handlers do not run.
//
// The important edge cases concern the absence and shape of the Origin header.
// A request with no Origin header is rejected with 403 by default; set
// Options.Optional to true to let such requests through, which is appropriate
// when the same routes must also serve same-origin navigations or non-browser
// clients that omit the header. An Origin that fails to parse as a URL is
// always rejected. Matching is exact and case-sensitive against the entries in
// Allow, so list each host exactly as browsers send it (for example
// "example.com" and, when a non-default port is used, "app.example.com:8443");
// the scheme is not compared, so both http and https origins for a listed host
// are accepted.
//
// Parity with the Node originals is behavioral rather than line-for-line. Like
// the common Express/Koa origin guards it compares the request Origin against a
// configured allowlist and blocks mismatches, but this port fixes the failure
// response as a plain 403 "Forbidden" rather than delegating to a
// user-supplied error handler, and it matches on host and hostname rather than
// on the full origin string including scheme. Callers needing scheme-strict or
// token-based semantics should layer additional middleware.
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
