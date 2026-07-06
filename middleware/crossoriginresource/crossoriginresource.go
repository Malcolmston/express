// Package crossoriginresource provides express middleware that sets the
// Cross-Origin-Resource-Policy (CORP) response header on every response. It is
// the Go analogue of helmet's crossOriginResourcePolicy feature (from the Node
// helmet package), packaged as a drop-in express.Handler. The one header it
// writes, in full, is "Cross-Origin-Resource-Policy".
//
// CORP lets a server declare which origins are allowed to include (embed, load,
// or otherwise consume) the resource it returns. It is the resource-side
// counterpart to Cross-Origin-Embedder-Policy: where COEP is a document opting
// into strict embedding, CORP is a resource opting into being embedded. Setting
// it protects your assets from being loaded by other sites — mitigating
// side-channel and speculative-execution attacks such as Spectre by keeping
// cross-origin resources out of an attacker's process — and provides the
// explicit opt-in that COEP: require-corp pages require of the resources they
// embed. Reach for this middleware on endpoints serving images, scripts, fonts,
// or APIs that should not be freely embeddable by arbitrary origins.
//
// Operationally the middleware is trivial and unconditional: it sits anywhere
// in the chain, reads nothing from the request, writes the
// Cross-Origin-Resource-Policy response header via res.Set, and always calls
// next() so the rest of the chain and the route handler run normally. It never
// short-circuits and applies to every response uniformly. Mount it globally
// with app.Use, or attach it only to the routes whose responses you want to
// protect.
//
// Behavior is driven by Options, whose zero value is usable and yields
// Cross-Origin-Resource-Policy: same-origin. The single field, Policy,
// overrides the header value; recognized values are "same-origin" (the default
// and the strictest, only same-origin documents may load the resource),
// "same-site" (any same-site origin may load it), and "cross-origin" (any
// origin may load it — appropriate for public CDN assets or APIs meant to be
// embedded everywhere). The value is written verbatim, so a custom or future
// token passes through unchanged, and there is no failure path. Take care with
// the default: same-origin will block legitimate cross-origin consumers of a
// public asset, so serve genuinely shareable resources with cross-origin.
//
// Compared with the helmet original, this port keeps the same single-header,
// same-origin-by-default contract but is intentionally minimal. Helmet
// validates the policy against a known set of tokens and can omit the header;
// here the header is always set and any non-empty Policy string is honored
// without validation, favoring flexibility over strict validation and leaving
// the choice of a valid token to the caller.
package crossoriginresource

import "github.com/malcolmston/express"

// Options configures the crossoriginresource middleware. The zero value is
// usable and yields Cross-Origin-Resource-Policy: same-origin.
type Options struct {
	// Policy overrides the header value (e.g. "same-origin", "same-site",
	// "cross-origin"). When empty, "same-origin" is used.
	Policy string
}

// New returns middleware that sets the Cross-Origin-Resource-Policy header.
func New(opts ...Options) express.Handler {
	value := "same-origin"
	if len(opts) > 0 && opts[0].Policy != "" {
		value = opts[0].Policy
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Cross-Origin-Resource-Policy", value)
		next()
	}
}
