// Package expectct provides express middleware that sets the Expect-CT response
// header, allowing sites to opt in to reporting and/or enforcement of
// Certificate Transparency requirements. It ports Helmet's expect-ct
// middleware, emitting the same max-age / enforce / report-uri directive set
// that helmet.expectCt() produced.
//
// Historically this header let a site instruct browsers to expect valid Signed
// Certificate Timestamps on its TLS certificates, reporting or rejecting
// connections that lacked them. Note that Expect-CT is effectively obsolete:
// Certificate Transparency is now enforced by default in modern browsers and
// the header is ignored by current versions, which is why Helmet itself removed
// it. This middleware exists for parity and for serving legacy clients; new
// deployments generally do not need it.
//
// The handler runs anywhere in the chain and only sets one response header. On
// every request it calls res.Set("Expect-CT", value) and then next()
// unconditionally, so it never inspects the request, never short-circuits, and
// leaves the status and body entirely to downstream handlers. Setting the
// header before next() ensures it is present when the response is eventually
// flushed. Mount it with app.Use, typically alongside other security headers.
//
// The header value is assembled once, at New time, from Options. It always
// begins with max-age=<seconds>; a negative MaxAge is clamped to 0, and the
// zero-value Options therefore yields the report-only, non-caching header
// "max-age=0". The "enforce" directive is appended when Enforce is true, and a
// report-uri="<url>" directive is appended when ReportURI is non-empty. The
// directives are joined with ", " in a fixed order (max-age, enforce,
// report-uri), so a fully configured value looks like
// max-age=86400, enforce, report-uri="https://example.com/report". The
// ReportURI is emitted verbatim inside double quotes without escaping or
// validation, so callers are responsible for supplying a well-formed URL.
//
// Regarding parity with the Node original: the directive names, ordering, and
// defaults follow Helmet's expect-ct implementation (max-age required,
// optional enforce, optional report-uri), and like Helmet the zero
// configuration emits max-age=0 in report-only mode. As with all Expect-CT
// deployments today, the practical caveat is that browsers no longer act on the
// header; this port reproduces the wire format but cannot change the fact that
// the mechanism has been deprecated in favor of built-in Certificate
// Transparency enforcement.
package expectct

import (
	"strconv"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the expectct middleware. The zero value is usable and
// yields a report-only header with max-age=0.
type Options struct {
	// MaxAge is the number of seconds the browser should cache the policy.
	MaxAge int

	// Enforce adds the "enforce" directive.
	Enforce bool

	// ReportURI, when non-empty, adds a report-uri directive.
	ReportURI string
}

// New returns middleware that sets the Expect-CT header.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	maxAge := o.MaxAge
	if maxAge < 0 {
		maxAge = 0
	}

	parts := []string{"max-age=" + strconv.Itoa(maxAge)}
	if o.Enforce {
		parts = append(parts, "enforce")
	}
	if o.ReportURI != "" {
		parts = append(parts, `report-uri="`+o.ReportURI+`"`)
	}
	value := strings.Join(parts, ", ")

	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Expect-CT", value)
		next()
	}
}
