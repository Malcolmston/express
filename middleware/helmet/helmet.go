// Package helmet bundles a sensible set of security-related HTTP response
// headers into a single express middleware, mirroring the popular Node.js
// Helmet defaults. It combines the behaviour of nosniff, frameguard, hsts,
// referrerpolicy, dnsprefetch, permittedcrossdomain, hidepoweredby, and
// originagentcluster.
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
