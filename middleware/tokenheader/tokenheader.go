// Package tokenheader provides generic middleware that validates an opaque
// token carried in a single, configurable request header. It is the express
// framework's Go take on the family of Node "API key" / "shared secret"
// header guards -- the hand-rolled middleware people write around headers such
// as X-API-Key, X-Auth-Token, or X-Session-Token, and the header mode of
// packages like express-api-key-auth -- packaged as a drop-in express.Handler.
// It is a close sibling of the bearerauth middleware, differing chiefly in
// that the header name is caller-chosen and no scheme prefix is parsed.
//
// Reach for this middleware when your credential does not live in the standard
// Authorization: Bearer form but in a custom header instead: internal
// service-to-service calls that pass a shared secret, webhook receivers that
// authenticate with a signing token in a vendor-specific header, or simple
// API-key gates on machine-consumed endpoints. Because the token is opaque to
// the middleware, it works equally well with random API keys, session
// identifiers, or signed tokens; the meaning and validity of the token are
// defined entirely by the caller-supplied Verify callback.
//
// Operationally the middleware belongs at the front of the chain, ahead of any
// handler that assumes an authenticated caller. On each request it reads the
// request header named by Options.Header via req.Get, treating an empty or
// unset header as no token, and passes whatever it finds to Options.Verify.
// The whole raw header value is used as the token; unlike bearerauth, no
// prefix is stripped and no surrounding whitespace is trimmed, so the Verify
// callback sees the header exactly as sent. When Verify returns true the
// middleware calls next() and the request proceeds unchanged; the token is not
// stashed on the request, so a handler needing the identity behind it should
// capture or look it up inside Verify.
//
// The request is short-circuited with res.Status(401).Send("Unauthorized") --
// and next() is never called -- whenever any of the following hold: Header is
// empty, the named header is missing or blank, Verify is nil, or Verify
// returns false. All of these failure modes collapse to the same bare 401 with
// no WWW-Authenticate challenge, so a caller cannot tell "no token" apart from
// "wrong token." Note that both Header and Verify are effectively required:
// with the zero-value Options{} every request is rejected, which fails closed
// rather than open.
//
// Security note: the token is a plaintext credential sent on every request, so
// this middleware must run over TLS, and comparisons inside Verify should use
// crypto/subtle.ConstantTimeCompare rather than == to resist timing attacks.
// Compared with heavier Node originals this port is intentionally minimal: it
// reads exactly one header, performs no scheme parsing, ships no token store
// or challenge header, and delegates every validity decision to Verify.
package tokenheader

import "github.com/malcolmston/express"

// Options configures the token-header middleware.
type Options struct {
	// Header names the request header carrying the token. Required.
	Header string
	// Verify reports whether the presented token is valid. Required.
	Verify func(token string) bool
}

// New returns middleware that reads the configured header and rejects the
// request with 401 when the token is missing or fails verification.
func New(opts Options) express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		token := ""
		if opts.Header != "" {
			token = req.Get(opts.Header)
		}
		if token == "" || opts.Verify == nil || !opts.Verify(token) {
			res.Status(401).Send("Unauthorized")
			return
		}
		next()
	}
}
