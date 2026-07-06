// Package crossoriginopener provides express middleware that sets the
// Cross-Origin-Opener-Policy (COOP) response header on every response. It is
// the Go analogue of helmet's crossOriginOpenerPolicy feature (from the Node
// helmet package), packaged as a drop-in express.Handler. The one header it
// writes, in full, is "Cross-Origin-Opener-Policy".
//
// COOP controls whether a document may share a browsing context group — and
// therefore a JavaScript reference via window.opener — with cross-origin
// documents it opens or is opened by. Setting it to same-origin severs that
// relationship for cross-origin windows, which both hardens the page against
// cross-window attacks such as tabnabbing and XS-Leaks and, together with a
// suitable Cross-Origin-Embedder-Policy, is one of the two headers required to
// make a document cross-origin isolated (unlocking SharedArrayBuffer and
// precise timers). Reach for this middleware to isolate your pages from the
// windows they interact with, or as part of enabling cross-origin isolation.
//
// Operationally the middleware is trivial and unconditional: it sits anywhere
// in the chain, reads nothing from the request, writes the
// Cross-Origin-Opener-Policy response header via res.Set, and always calls
// next() so the remaining handlers and the route run normally. It never
// short-circuits and applies to every response uniformly. Mount it globally
// with app.Use to cover the whole application.
//
// Behavior is driven by Options, whose zero value is usable and yields
// Cross-Origin-Opener-Policy: same-origin. The single field, Policy, overrides
// the header value; recognized values are "same-origin" (the default and the
// strictest, isolating the document from all cross-origin windows),
// "same-origin-allow-popups" (the document keeps a reference to popups it opens
// but is isolated from documents that open it), and "unsafe-none" (the browser
// default, disabling the protection). The value is written verbatim, so any
// custom or future token passes through unchanged, and there is no failure
// path. Note that same-origin can break integrations that rely on
// window.opener or postMessage across origins (OAuth popups, payment flows,
// federated login), so choose same-origin-allow-popups when such flows are
// required.
//
// Compared with the helmet original, this port keeps the same single-header,
// same-origin-by-default contract but is intentionally minimal. Helmet
// validates the policy against a known set of tokens and can omit the header;
// here the header is always set and any non-empty Policy string is honored
// without validation, trading strict validation for flexibility and placing the
// responsibility for a valid token on the caller.
package crossoriginopener

import "github.com/malcolmston/express"

// Options configures the crossoriginopener middleware. The zero value is usable
// and yields Cross-Origin-Opener-Policy: same-origin.
type Options struct {
	// Policy overrides the header value (e.g. "same-origin",
	// "same-origin-allow-popups", "unsafe-none"). When empty, "same-origin" is
	// used.
	Policy string
}

// New returns middleware that sets the Cross-Origin-Opener-Policy header.
func New(opts ...Options) express.Handler {
	value := "same-origin"
	if len(opts) > 0 && opts[0].Policy != "" {
		value = opts[0].Policy
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Cross-Origin-Opener-Policy", value)
		next()
	}
}
