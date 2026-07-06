// Package helmet bundles a sensible set of security-related HTTP response
// headers into a single express middleware, mirroring the popular Node.js
// Helmet defaults. It is a stdlib-only port of helmet's default header set to
// this Express-style framework, composing what upstream ships as several
// independent sub-middleware into one handler so a single app.Use hardens every
// response.
//
// Reach for helmet as a baseline security layer on any web application: it is
// the quickest way to turn on a batch of widely recommended, low-risk response
// headers without wiring each one individually. Each header addresses a
// specific attack surface — MIME sniffing, clickjacking, protocol downgrade,
// referrer leakage, DNS prefetch tracking, cross-domain policy abuse, stack
// fingerprinting, and cross-origin process isolation — and applying them
// together closes several classes of common web vulnerabilities at once. If you
// need finer control over any single header you can still use the dedicated
// sibling middleware (hsts, frameguard, hidepoweredby, dnsprefetch, and so on),
// which this package mirrors.
//
// This handler composes the behavior of several sub-middleware into one call.
// In one pass it writes X-Content-Type-Options: nosniff (nosniff), X-Frame-
// Options (frameguard), Strict-Transport-Security (hsts), Referrer-Policy
// (referrerpolicy), X-DNS-Prefetch-Control: off (dnsprefetch), X-Permitted-
// Cross-Domain-Policies: none (permittedcrossdomain), and Origin-Agent-Cluster:
// ?1 (originagentcluster). It also incorporates hidepoweredby by registering a
// res.OnBeforeWrite hook that deletes the X-Powered-By header just before the
// response is committed, so the banner is stripped regardless of when a
// downstream handler adds it.
//
// Chain placement and control flow: register helmet early, before your route
// handlers, so the headers are present on every response including errors. The
// handler reads no request state; it only writes response headers via res.Set,
// installs the one before-write hook, and then unconditionally calls next(). It
// never short-circuits, so it is transparent to routing and never touches the
// response body. All header values are computed once when New is called and
// reused for every request.
//
// Behavior and defaults are driven by the optional Options argument (the zero
// value applies every default). X-Content-Type-Options, X-DNS-Prefetch-Control,
// X-Permitted-Cross-Domain-Policies, and Origin-Agent-Cluster are fixed.
// FrameguardAction selects the X-Frame-Options value and defaults to SAMEORIGIN
// (DENY is the only other accepted value, matched case-insensitively; any other
// input falls back to SAMEORIGIN). ReferrerPolicy defaults to no-referrer.
// The HSTS directive defaults to max-age=15552000 (180 days); HSTSMaxAge
// overrides the seconds and a negative value emits max-age=0, while
// HSTSIncludeSubDomains appends includeSubDomains. Remember that the
// Strict-Transport-Security header only takes effect over HTTPS.
//
// Parity with the Node original: helmet enables this same set of headers by
// default and defaults X-Frame-Options to SAMEORIGIN and Referrer-Policy to
// no-referrer, matching this port. Intentional differences: upstream helmet no
// longer enables its legacy Content-Security-Policy, X-XSS-Protection, or
// Expect-CT middleware by default, and this port likewise omits them (use the
// sibling csp and expectct packages if you need them). helmet's preload HSTS
// directive and its richer per-middleware options are also not surfaced here;
// compose the dedicated middleware when you need that granularity.
package helmet

import (
	"strconv"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the helmet middleware. The zero value is usable and
// applies all defaults.
type Options struct {
	// HSTSMaxAge overrides the Strict-Transport-Security max-age in seconds.
	// When zero, the default of 15552000 (180 days) is used. A negative value
	// sends max-age=0.
	HSTSMaxAge int

	// HSTSIncludeSubDomains adds the includeSubDomains directive to HSTS.
	HSTSIncludeSubDomains bool

	// FrameguardAction overrides X-Frame-Options ("SAMEORIGIN" or "DENY").
	// When empty, "SAMEORIGIN" is used.
	FrameguardAction string

	// ReferrerPolicy overrides Referrer-Policy. When empty, "no-referrer" is
	// used.
	ReferrerPolicy string
}

const defaultHSTSMaxAge = 15552000

// New returns middleware that sets helmet's bundle of default security headers.
// It accepts an optional Options value (only the first is used; omit it to take
// all defaults) and precomputes the X-Frame-Options, Referrer-Policy, and
// Strict-Transport-Security values once. The returned handler writes the seven
// security headers on every request, registers a before-write hook that deletes
// X-Powered-By, and always calls next without short-circuiting.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	frame := "SAMEORIGIN"
	if strings.EqualFold(o.FrameguardAction, "DENY") {
		frame = "DENY"
	}

	referrer := "no-referrer"
	if o.ReferrerPolicy != "" {
		referrer = o.ReferrerPolicy
	}

	maxAge := o.HSTSMaxAge
	switch {
	case maxAge == 0:
		maxAge = defaultHSTSMaxAge
	case maxAge < 0:
		maxAge = 0
	}
	hstsParts := []string{"max-age=" + strconv.Itoa(maxAge)}
	if o.HSTSIncludeSubDomains {
		hstsParts = append(hstsParts, "includeSubDomains")
	}
	hstsValue := strings.Join(hstsParts, "; ")

	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Content-Type-Options", "nosniff")
		res.Set("X-Frame-Options", frame)
		res.Set("Strict-Transport-Security", hstsValue)
		res.Set("Referrer-Policy", referrer)
		res.Set("X-DNS-Prefetch-Control", "off")
		res.Set("X-Permitted-Cross-Domain-Policies", "none")
		res.Set("Origin-Agent-Cluster", "?1")

		// hidepoweredby: strip the X-Powered-By header just before the
		// response is committed.
		res.OnBeforeWrite(func() {
			res.Writer.Header().Del("X-Powered-By")
		})

		next()
	}
}
