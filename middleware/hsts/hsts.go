// Package hsts provides middleware that sets the HTTP
// Strict-Transport-Security (HSTS) response header, instructing conforming
// browsers to access the site only over HTTPS. It is a stdlib-only port of the
// Node.js "hsts" package (also exposed as helmet.hsts) to this Express-style
// framework, and produces the same header values for equivalent options.
//
// Use this middleware on any site served over TLS to defend against protocol
// downgrade attacks and cookie hijacking. Once a browser has seen the
// Strict-Transport-Security header it will, for the lifetime of max-age,
// automatically upgrade http:// requests to https:// and refuse to connect if
// the certificate is invalid, closing the window in which a network attacker
// could intercept a plaintext request. It does not itself perform any redirect;
// pair it with a force-HTTPS redirect for the very first, still-plaintext hit.
//
// In the middleware chain it should sit early, typically among the other
// security-header middleware, so the header is present on every response
// including error responses. The handler reads no request state: on each call
// it writes a single Strict-Transport-Security response header via res.Set and
// then unconditionally invokes next, so it never short-circuits and never
// interferes with the body. The header value is computed once when New is
// called and reused for every request, so per-request overhead is a single map
// write.
//
// The value is assembled from the configured Options as
// "max-age=<seconds>" optionally followed by "; includeSubDomains" and
// "; preload". MaxAge defaults to DefaultMaxAge (15552000 seconds, 180 days)
// when zero; a negative MaxAge is clamped to zero, which tells browsers to
// forget the policy. Note that browsers only honor the header on secure
// (HTTPS) connections and ignore it over plain HTTP, and that submitting a
// domain to a browser preload list (via the preload directive) is effectively
// irreversible on normal timescales, so enable Preload only when you are
// certain every subdomain will support HTTPS indefinitely.
//
// Parity with the Node original: the assembled directive string and the
// max-age defaulting/clamping match helmet's hsts. This port intentionally
// omits the deprecated Node "setIf" predicate and the legacy "maxAge" alias
// spellings; configure behavior through the typed Options struct instead. Only
// the Strict-Transport-Security header is managed here — for the full helmet
// bundle use the sibling helmet package.
package hsts

import (
	"strconv"
	"strings"

	"github.com/malcolmston/express"
)

// DefaultMaxAge is the default max-age in seconds (180 days).
const DefaultMaxAge = 15552000

// Options configures the HSTS middleware. The zero value is usable and uses
// DefaultMaxAge with neither includeSubDomains nor preload.
type Options struct {
	// MaxAge is the number of seconds browsers should remember to prefer
	// HTTPS. When zero, DefaultMaxAge is used. Use a negative value to send
	// max-age=0.
	MaxAge int

	// IncludeSubDomains adds the includeSubDomains directive.
	IncludeSubDomains bool

	// Preload adds the preload directive for inclusion in browser preload
	// lists.
	Preload bool
}

// New returns middleware that sets the Strict-Transport-Security header.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	maxAge := o.MaxAge
	switch {
	case maxAge == 0:
		maxAge = DefaultMaxAge
	case maxAge < 0:
		maxAge = 0
	}

	parts := []string{"max-age=" + strconv.Itoa(maxAge)}
	if o.IncludeSubDomains {
		parts = append(parts, "includeSubDomains")
	}
	if o.Preload {
		parts = append(parts, "preload")
	}
	value := strings.Join(parts, "; ")

	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Strict-Transport-Security", value)
		next()
	}
}
