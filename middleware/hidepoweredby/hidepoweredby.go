// Package hidepoweredby provides middleware that removes the X-Powered-By
// response header (or replaces it with a decoy value) so the server does not
// advertise the technology stack it runs on. It is a stdlib-only port of the
// Node.js "hide-powered-by" package (exposed in helmet as hidePoweredBy) to
// this Express-style framework.
//
// Reach for this middleware whenever you want to reduce information leakage in
// your response headers. By default many frameworks — including this one, whose
// "x-powered-by" application setting is enabled by default — advertise
// themselves with an X-Powered-By header. That value gives attackers a free
// hint about the server software and version to target, so security hardening
// guides routinely recommend stripping it. Supplying a decoy value instead
// (for example pretending to be a different stack) can additionally mislead
// automated scanners.
//
// In the chain the middleware should run before the response is written; the
// simplest placement is an early app.Use. Rather than deleting the header
// immediately, New registers a res.OnBeforeWrite hook that fires just before
// the response headers are committed. This is important because the framework
// (and downstream handlers) may add or re-add X-Powered-By after this
// middleware runs, so acting at write time guarantees the final header state
// wins regardless of ordering. The handler always calls next and never
// short-circuits, so it is transparent to routing and the response body.
//
// Behavior is controlled by Options.SetTo. When SetTo is empty (the zero
// value) the hook deletes X-Powered-By outright. When SetTo is non-empty the
// hook sets X-Powered-By to that decoy string, overriding any value the
// framework would otherwise emit. Note that if you only want deletion, the
// application's own "x-powered-by" setting must also be considered: this
// framework re-adds a default "Express" value when that setting is enabled, and
// the before-write hook here removes it, so the two interact predictably at
// commit time.
//
// Parity with the Node original: the effective outcome matches
// hide-powered-by — remove the header by default, or replace it with a chosen
// value via setTo. Because Go's net/http already omits any framework banner
// unless one is explicitly added, this port implements the behavior with a
// write-time header hook rather than patching a response prototype. It manages
// only the X-Powered-By header and no other security headers.
package hidepoweredby

import "github.com/malcolmston/express"

// Options configures the hidepoweredby middleware. The zero value is usable and
// removes the X-Powered-By header entirely.
type Options struct {
	// SetTo, when non-empty, replaces X-Powered-By with this decoy value
	// instead of removing the header.
	SetTo string
}

// New returns middleware that hides or spoofs the X-Powered-By header. The
// header is adjusted via a before-write hook so it takes effect regardless of
// when downstream handlers commit the response.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.OnBeforeWrite(func() {
			if o.SetTo != "" {
				res.Writer.Header().Set("X-Powered-By", o.SetTo)
			} else {
				res.Writer.Header().Del("X-Powered-By")
			}
		})
		next()
	}
}
