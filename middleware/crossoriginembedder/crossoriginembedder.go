// Package crossoriginembedder provides express middleware that sets the
// Cross-Origin-Embedder-Policy (COEP) response header on every response. It is
// the Go analogue of helmet's crossOriginEmbedderPolicy feature (from the Node
// helmet package), packaged as a drop-in express.Handler. The one header it
// writes, in full, is "Cross-Origin-Embedder-Policy".
//
// COEP controls whether a document is allowed to load cross-origin resources
// that have not explicitly opted in to being embedded. Its main purpose is to
// help a page become "cross-origin isolated": when a document is served with
// both COEP: require-corp (this middleware) and a suitable
// Cross-Origin-Opener-Policy, the browser grants access to powerful features
// that are otherwise disabled for security reasons, such as SharedArrayBuffer,
// high-resolution timers, and precise performance measurement. Reach for this
// middleware when you need cross-origin isolation, or simply when you want to
// prevent a page from silently embedding cross-origin resources that did not
// consent to it.
//
// Operationally the middleware is trivial and unconditional: it sits anywhere
// in the chain, reads nothing from the request, writes the
// Cross-Origin-Embedder-Policy response header via res.Set, and always calls
// next() so the rest of the chain and the route handler run normally. It never
// short-circuits, never inspects the method or path, and applies to every
// response uniformly. Mount it globally with app.Use to cover the whole
// application.
//
// Behavior is driven by Options, whose zero value is usable and yields
// Cross-Origin-Embedder-Policy: require-corp. The single field, Policy,
// overrides the header value; recognized values are "require-corp" (the
// default and the strictest), "credentialless" (cross-origin subresources are
// fetched without credentials and thus need no explicit CORP opt-in), and
// "unsafe-none" (the browser default, effectively disabling the protection).
// The value is written verbatim, so a custom or future token is passed through
// unchanged. There is no failure path. Be aware that require-corp is a strict
// policy: any cross-origin image, script, iframe, or font that lacks a
// permissive Cross-Origin-Resource-Policy or CORS response will be blocked, so
// enable it only after auditing the resources your pages embed.
//
// Compared with the helmet original, this port keeps the same single-header,
// require-corp-by-default contract but is intentionally minimal. Helmet accepts
// a policy option restricted to a known set of tokens and can disable the
// header entirely; here the header is always set and any non-empty Policy
// string is honored without validation, which is more flexible but shifts the
// responsibility for supplying a valid token onto the caller.
package crossoriginembedder

import "github.com/malcolmston/express"

// Options configures the crossoriginembedder middleware. The zero value is
// usable and yields Cross-Origin-Embedder-Policy: require-corp.
type Options struct {
	// Policy overrides the header value (e.g. "require-corp",
	// "credentialless", "unsafe-none"). When empty, "require-corp" is used.
	Policy string
}

// New returns middleware that sets the Cross-Origin-Embedder-Policy header.
func New(opts ...Options) express.Handler {
	value := "require-corp"
	if len(opts) > 0 && opts[0].Policy != "" {
		value = opts[0].Policy
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Cross-Origin-Embedder-Policy", value)
		next()
	}
}
