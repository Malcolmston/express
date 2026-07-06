// Package xssfilter provides express middleware that sets the legacy
// X-XSS-Protection response header on every response. It is the express analogue
// of Helmet's xssFilter/X-XSS-Protection helper for Node, packaged as a small
// middleware so an application can control this one header in a single place at
// the front of its chain.
//
// X-XSS-Protection was a header that instructed a browser's built-in reflected-
// XSS auditor how to behave. Modern security guidance — and the default here —
// is to turn that auditor off by sending the value "0". The heuristic auditors
// in older browsers were themselves a source of vulnerabilities and subtle
// content-corruption bugs, every major engine has since removed the feature, and
// real cross-site-scripting defense belongs in a Content-Security-Policy and in
// correct output encoding. Emitting "0" documents that intent explicitly and
// suppresses any residual legacy behavior.
//
// Mechanically New returns a handler that, on each request, calls res.Set to
// write the X-XSS-Protection header and then calls next() to continue the chain.
// It never inspects the request, reads the response body, or short-circuits, so
// it is transparent to routing and safe to mount anywhere — though placing it
// early with app.Use ensures the header is present even on responses produced by
// later error handlers. The header is set, not appended, so a single value is
// emitted per response.
//
// The value is configurable through Options. The zero value is usable and yields
// X-XSS-Protection: 0; supplying a non-empty Options.Value overrides it, which is
// the escape hatch for the rare deployment that must reproduce a legacy directive
// such as "1; mode=block" for a specific client. An empty Value is treated as
// "not set" and falls back to the "0" default rather than emitting an empty
// header.
//
// Compared with the Node/Helmet original this port keeps the same defensive
// default of disabling the auditor and the same single-header responsibility,
// while trimming the option surface to one Value override rather than Helmet's
// boolean/mode configuration object. It reproduces the header's observable
// effect exactly; the difference is only in how the desired value is expressed.
package xssfilter

import "github.com/malcolmston/express"

// Options configures the xssfilter middleware. The zero value is usable and
// yields X-XSS-Protection: 0.
type Options struct {
	// Value overrides the header value. When empty, "0" is used.
	Value string
}

// New returns middleware that sets the X-XSS-Protection header.
func New(opts ...Options) express.Handler {
	value := "0"
	if len(opts) > 0 && opts[0].Value != "" {
		value = opts[0].Value
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-XSS-Protection", value)
		next()
	}
}
