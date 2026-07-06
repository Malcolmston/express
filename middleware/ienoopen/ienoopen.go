// Package ienoopen provides middleware that sets the X-Download-Options
// response header to the single value "noopen". It is a direct port of the
// Helmet "ienoopen" / X-Download-Options component (the same behavior Helmet
// exposes today through helmet.ieNoOpen()), reproducing that header in a
// stdlib-only Express-style stack for Go.
//
// The header exists to defend a very specific legacy behavior in Internet
// Explorer 8 and later. Historically those browsers offered an "Open" button
// in the file-download dialog that executed a downloaded file directly in the
// security context of the originating site. For an application that serves
// user-supplied or otherwise untrusted downloads, that behavior allowed an
// attacker-controlled file to run as if it came from your origin. Sending
// X-Download-Options: noopen removes the "Open" option, forcing users to save
// the file to disk before opening it. Use this middleware when your service
// hands out downloadable content and you still care about hardening older IE
// clients; on modern browsers the header is simply ignored and harmless.
//
// The middleware is a header-only, non-terminating handler and is meant to sit
// early in the chain, typically registered globally with app.Use so it applies
// to every route. On each request it reads no request state and writes exactly
// one response header via res.Set("X-Download-Options", "noopen"), then always
// calls next() to continue processing. It never inspects the method, path, or
// body and never short-circuits, so it composes cleanly with routing,
// authentication, and body-producing handlers that run after it.
//
// Semantically there is nothing to configure: New takes no options, the value
// is always the constant "noopen", and the header is set unconditionally
// (including on error responses produced downstream, as long as headers have
// not already been flushed). Because it uses res.Set, a later handler can still
// override or delete the header if it needs to; the middleware makes no attempt
// to lock the value. There is no short-circuit path and no error return, so it
// cannot fail on its own.
//
// Parity with the Node original is exact in the part that matters: the emitted
// header name and value are identical to Helmet's, which is the entire contract
// of this feature. Helmet's implementation is likewise a fixed-string header
// setter with no options, so this port has full behavioral parity. The only
// differences are idiomatic — it returns an express.Handler and writes through
// the Response abstraction rather than a Node res.setHeader call.
package ienoopen

import "github.com/malcolmston/express"

// New returns middleware that sets X-Download-Options: noopen.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Download-Options", "noopen")
		next()
	}
}
