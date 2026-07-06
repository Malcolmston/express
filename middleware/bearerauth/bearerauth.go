// Package bearerauth provides middleware that authenticates requests using an
// opaque bearer token supplied in the Authorization header. It is the express
// framework's Go analogue of the "Bearer" scheme handled by Node middleware
// such as express-bearer-token and passport-http-bearer, packaged as a drop-in
// express.Handler that challenges callers with a WWW-Authenticate header and
// rejects anything without a valid token.
//
// Reach for this middleware when you want to gate routes behind a token instead
// of a username/password pair: API endpoints consumed by scripts and services,
// mobile or SPA back ends that already mint access tokens, webhook receivers,
// or any machine-to-machine call that carries a credential in the standard
// "Authorization: Bearer <token>" form. Because the token is opaque to the
// middleware, it works equally well with random session identifiers, API keys,
// or signed tokens such as JWTs; the meaning of the token is entirely defined
// by the caller-supplied Verify callback.
//
// Operationally the middleware belongs at the front of the chain, before any
// handler that assumes an authenticated caller. On each request it reads the
// Authorization request header, requires a case-insensitive "Bearer " prefix,
// trims surrounding whitespace from the remainder, and passes the resulting
// token to Options.Verify. Verify is the single source of truth for validity:
// when it returns true the middleware calls next() and the request proceeds
// unchanged, and when it returns false the request is stopped. The token is not
// stashed on the request, so a handler that needs the identity behind the token
// should capture or look it up inside Verify.
//
// When the header is missing, does not use the Bearer scheme, carries an empty
// token, or Verify returns false (or Verify is nil), the request is
// short-circuited: the middleware sets a "WWW-Authenticate: Bearer" response
// header and writes a 401 Unauthorized body, and next() is never called. Every
// failure mode collapses to the same bare "Bearer" challenge and identical 401
// so that a caller cannot distinguish "no token" from "wrong token"; the
// middleware intentionally does not emit error or error_description parameters.
//
// Security note: a bearer token is a plaintext credential presented on every
// request, so this middleware must always run over TLS. Comparisons performed
// inside Verify are the caller's responsibility; to resist timing attacks,
// compare secrets with crypto/subtle.ConstantTimeCompare rather than ==.
// Compared with the Node originals, this port keeps the challenge-and-reject
// contract but is deliberately minimal: it does not parse tokens out of query
// strings or form bodies, ships no token store, and delegates every validity
// decision to Verify.
package bearerauth

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the bearer authentication middleware.
type Options struct {
	// Verify reports whether the presented token is valid. It is required.
	Verify func(token string) bool
}

// New returns middleware that reads an "Authorization: Bearer <token>" header
// and rejects the request with 401 when the token is missing or invalid.
func New(opts Options) express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		token, ok := extractToken(req.Get("Authorization"))
		if !ok || opts.Verify == nil || !opts.Verify(token) {
			res.Set("WWW-Authenticate", "Bearer")
			res.Status(401).Send("Unauthorized")
			return
		}
		next()
	}
}

// extractToken pulls the token out of a "Bearer <token>" authorization header.
func extractToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}
