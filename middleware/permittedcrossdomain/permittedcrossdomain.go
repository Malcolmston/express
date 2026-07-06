// Package permittedcrossdomain provides middleware that sets the
// X-Permitted-Cross-Domain-Policies response header on every response,
// controlling whether Adobe clients honor cross-domain policy files served by
// the site. It is a port of Helmet's crossOriginResourcePolicy sibling — the
// permittedCrossDomainPolicies middleware — which emits this single header with
// a configurable policy value.
//
// The header governs legacy Adobe technologies (Flash Player and Acrobat) that
// look for a crossdomain.xml or similar policy file to decide whether they may
// load data from the site on behalf of another origin. Setting the header to
// "none" instructs those clients that no policy files are permitted anywhere on
// the domain, closing a historical cross-origin data-theft vector. Use it as a
// low-cost hardening measure on any site; it is harmless on modern browsers,
// which ignore the header, and meaningful only to the shrinking set of Adobe
// runtimes that still consult it.
//
// Mechanically the middleware is trivial and stateless: for each request it
// sets X-Permitted-Cross-Domain-Policies to the configured value and then calls
// next to continue the chain. It never reads request state, never
// short-circuits, and never writes a body, so it composes cleanly with other
// handlers. Register it early via app.Use so the header is present on every
// response, including error responses, as long as it runs before the headers
// are flushed.
//
// The single option, Options.Policy, selects the header value. When it is empty
// — including when New is called with no Options at all — the middleware
// defaults to the most restrictive value, "none". Other accepted values are
// "master-only" (only the master policy file at the domain root is honored),
// "by-content-type" (only policy files served with the correct content type),
// and "all" (any policy file is honored, the least safe choice). The package
// does not validate the string, so an unknown value is sent verbatim; supplying
// one of the standard tokens is the caller's responsibility.
//
// Parity with the Node original is exact for the wire behavior: like
// helmet.permittedCrossDomainPolicies() this package sets exactly one header
// and defaults it to "none", overridable by a single policy option. It does not
// reproduce Helmet's runtime validation of the policy string; the observable
// result — the header and its value on every response — matches the Node
// middleware for all valid inputs.
package permittedcrossdomain

import "github.com/malcolmston/express"

// Options configures the permittedcrossdomain middleware. The zero value is
// usable and yields X-Permitted-Cross-Domain-Policies: none.
type Options struct {
	// Policy overrides the header value (e.g. "none", "master-only",
	// "by-content-type", "all"). When empty, "none" is used.
	Policy string
}

// New returns middleware that sets the X-Permitted-Cross-Domain-Policies header.
func New(opts ...Options) express.Handler {
	value := "none"
	if len(opts) > 0 && opts[0].Policy != "" {
		value = opts[0].Policy
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Permitted-Cross-Domain-Policies", value)
		next()
	}
}
