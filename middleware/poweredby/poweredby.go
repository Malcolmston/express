// Package poweredby provides express middleware that sets the X-Powered-By
// response header to a configurable value. It is the deliberate inverse of the
// hidepoweredby middleware (and of Helmet's hidePoweredBy feature): where those
// strip the header to avoid advertising the technology stack, poweredby writes
// an explicit, chosen value so branding is under your control rather than left
// to a framework default.
//
// In stock Express, the app implicitly emits "X-Powered-By: Express" on every
// response unless the "x-powered-by" application setting is disabled. This
// middleware ports that behaviour to the stdlib-only express port: rather than
// relying on any implicit framework default, it sets the header to Value on
// each response. This makes the branding value explicit and testable, and lets
// applications present a custom product string (for example "MyApp/2.0") or
// suppress the header entirely.
//
// Use it near the top of the middleware chain, before route handlers, so the
// header is present on every response regardless of which handler ultimately
// writes the body. On each request the middleware inspects the resolved value:
// when non-empty it calls res.Set("X-Powered-By", value); when empty it removes
// any existing X-Powered-By header via res.Writer.Header().Del. It never
// short-circuits the chain and always invokes next() to continue processing.
//
// The value is resolved once when New is called. With no Options the constant
// DefaultValue ("Express") is used, matching stock Express. Supplying
// Options{Value: ""} explicitly is a supported way to delete the header (behave
// like hidepoweredby's removal mode); supplying any other string sets that
// literal value. Because the header is written on the response writer's header
// map, the middleware must run before the response is committed, which is the
// normal case for a leading middleware.
//
// Compared with the Node original, parity is intentionally narrow: Express only
// exposes an on/off application setting and a fixed "Express" string, whereas
// this port additionally allows an arbitrary branding value through Options.
// Because advertising the server technology can aid attackers during
// reconnaissance, security-conscious deployments typically prefer hidepoweredby
// (removal) over poweredby; this middleware exists for the cases where an
// explicit, custom X-Powered-By value is genuinely wanted.
package poweredby

import "github.com/malcolmston/express"

// DefaultValue is used when no Value option is supplied.
const DefaultValue = "Express"

// Options configures the middleware.
type Options struct {
	// Value is written as the X-Powered-By header.
	Value string
}

// New returns middleware that sets the X-Powered-By response header. This
// overrides any framework default so branding can be customized (or hidden by
// supplying an empty-but-explicit value, in which case the header is removed).
func New(opts ...Options) express.Handler {
	value := DefaultValue
	if len(opts) > 0 {
		value = opts[0].Value
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		if value == "" {
			res.Writer.Header().Del("X-Powered-By")
		} else {
			res.Set("X-Powered-By", value)
		}
		next()
	}
}
