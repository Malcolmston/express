// Package nosniff provides middleware that sets the response header
// X-Content-Type-Options: nosniff on every request. It is the express port of
// Helmet's noSniff middleware (and the equivalent dontSniffMimetype behavior of
// the Node lusca and koa-helmet families), which in turn wraps the single
// header that all modern browsers honor to disable content-type sniffing.
//
// Use it whenever you serve responses whose bytes could be misinterpreted if a
// browser ignored the declared Content-Type: user-uploaded files, JSON or text
// endpoints that an attacker might coax a browser into executing as HTML or
// JavaScript, and any response where "the server said so" should be the final
// word on the media type. It is cheap enough to enable globally and is part of
// most sensible security baselines, which is why Helmet turns it on by default.
//
// Mechanically the middleware is trivial and side-effect free with respect to
// the body: it calls res.Set("X-Content-Type-Options", "nosniff") and then
// immediately invokes next() to continue the chain. It writes exactly one
// header, reads no request state, and never short-circuits or terminates the
// response, so its position in the chain only matters to the extent that it
// must run before the headers are flushed. Registering it early via app.Use is
// the usual choice so that the header is present on both normal and error
// responses.
//
// The important semantics are that "nosniff" is the only valid value for this
// header, so there are no options to configure and no defaults to reason about;
// the middleware is intentionally a constant. Setting it tells browsers to
// enforce the declared Content-Type for script and style contexts (blocking, for
// example, a script tag that points at a text/plain resource) and to apply MIME
// type checking, which closes a class of drive-by and content-confusion attacks.
// If a later handler overwrites X-Content-Type-Options the last write wins, but
// nothing in this package removes or weakens the header once set.
//
// Parity with the Node original is exact for the behavior that matters: like
// helmet.noSniff() this package emits a fixed X-Content-Type-Options: nosniff and
// nothing else. It does not attempt to reproduce Helmet's broader bundle of
// headers (see the helmet middleware for that), and because Go's net/http handles
// header casing canonically there is no configuration surface to diverge on.
package nosniff

import "github.com/malcolmston/express"

// New returns middleware that sets X-Content-Type-Options: nosniff.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Content-Type-Options", "nosniff")
		next()
	}
}
