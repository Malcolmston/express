// Package scopecheck provides middleware that authorizes a request only when
// the caller holds every one of a configured set of scopes. It is the express
// framework's Go analogue of the OAuth 2.0 / OpenID Connect scope enforcement
// that Node services layer on top of express-oauth2-jwt-bearer, express-jwt, or
// Passport strategies: an earlier stage validates the access token and exposes
// its granted scopes, and this middleware decides whether those scopes cover
// what the endpoint demands.
//
// Reach for this middleware to protect an API route whose access token must
// carry specific permissions — for example requiring both "read" and "write"
// on a mutating endpoint, or "profile" and "email" before returning user data.
// The scopes are plain strings compared for equality, so the middleware does
// not care whether they arrive from an OAuth token's "scope" claim, an API-key
// record, or a session; the Getter callback is the sole bridge between the
// request and the scope list.
//
// Operationally the middleware belongs after token validation and after
// whatever step attaches the granted scopes to the request, but before the
// protected handler. On each request it calls Options.Getter to collect the
// scopes the caller holds and loads them into a set for O(1) lookup, then walks
// Options.Required. If every required scope is present it calls next() exactly
// once and the request proceeds unchanged; the guard writes nothing to the
// response and adds nothing to the request on the success path.
//
// The check is a logical AND: the caller must hold all of Options.Required,
// which is what distinguishes it from the sibling rolecheck package, whose
// check is a logical OR over the acceptable roles. Matching is case-sensitive
// and exact. As soon as one required scope is missing the request is
// short-circuited with 403 Forbidden and a plain "Forbidden" body, and next()
// is never called. A nil Options.Getter is tolerated: it yields an empty scope
// set, so any non-empty Options.Required rejects the request, while an empty
// Options.Required trivially passes because there is nothing left to satisfy.
//
// Compared with the full OAuth scope middlewares it stands in for, this port is
// deliberately minimal. It performs no token parsing or signature verification,
// understands no scope hierarchies, wildcards, or space-delimited scope strings
// (the caller must split those before returning them from Getter), and offers
// no control over the status code or body — every denial collapses to a bare
// 403. Extracting and interpreting scopes is entirely the caller's job.
package scopecheck

import "github.com/malcolmston/express"

// Options configures the scope-check middleware.
type Options struct {
	// Required lists the scopes that must all be present. Required.
	Required []string
	// Getter extracts the scopes associated with a request. Required.
	Getter func(req *express.Request) []string
}

// New returns middleware that responds with 403 unless every required scope is
// present on the request.
func New(opts Options) express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		var have []string
		if opts.Getter != nil {
			have = opts.Getter(req)
		}
		set := make(map[string]struct{}, len(have))
		for _, s := range have {
			set[s] = struct{}{}
		}
		for _, want := range opts.Required {
			if _, ok := set[want]; !ok {
				res.Status(403).Send("Forbidden")
				return
			}
		}
		next()
	}
}
