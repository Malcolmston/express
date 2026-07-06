// Package forcessl provides middleware that redirects insecure HTTP requests to
// their HTTPS equivalent, preserving the host, path, and query string. It is a
// stdlib-only port of the Node "express-force-ssl" (and the closely related
// "express-sslify") middleware: when a request arrives over plain HTTP it is
// bounced to the same URL under the https scheme so that browsers and clients
// upgrade to an encrypted connection.
//
// Use this middleware at the very top of the chain on any deployment that must
// never serve content over cleartext HTTP — for example a public site behind a
// load balancer or reverse proxy that also terminates or forwards TLS. Placing
// it first ensures the redirect happens before any handler runs, so no
// application logic executes and no cookies or bodies are processed on the
// insecure request. It pairs naturally with an HSTS header set by a separate
// security middleware once the client is on HTTPS.
//
// On each request the middleware inspects req.Secure(). A request is considered
// secure when it either arrived over a TLS connection or carries the
// X-Forwarded-Proto: https header (the common signal from a TLS-terminating
// proxy) — both cases are collapsed into req.Secure() by the framework. If the
// request is already secure, or if the middleware is disabled, it simply calls
// next() and the chain continues untouched. Otherwise it short-circuits the
// chain: it does not call next() and instead writes a 301 (Moved Permanently)
// redirect to "https://" + req.Raw.Host + req.OriginalURL(), reusing the
// original host and the full original request URI (path plus query string) so
// deep links and query parameters survive the upgrade.
//
// The single option, Enabled, gates the whole behavior. Because a bool's zero
// value is false, New defaults Enabled to true when Options are omitted
// entirely, so New() with no arguments enforces HTTPS; to turn the redirect off
// (for instance in local development) pass Options{Enabled: false} explicitly,
// which makes every request pass straight through. A 301 is used rather than a
// 302 so that browsers and intermediaries cache the upgrade; be aware that 301s
// are aggressively cached, so misconfiguring the redirect can be sticky on
// clients. Because req.Secure() trusts X-Forwarded-Proto, only run this behind a
// proxy you control that strips client-supplied forwarding headers, or a client
// could spoof "https" and bypass the redirect.
//
// Compared with the Node original this port keeps the core scheme-upgrade
// behavior but is deliberately minimal: there are no options for a custom
// redirect status, an explicit hostname override, a trust-proxy toggle, or port
// remapping. The status is fixed at 301 and proxy trust is delegated to the
// framework's req.Secure() implementation. If you need those knobs, wrap or fork
// this middleware.
package forcessl

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the force-SSL middleware.
type Options struct {
	// Enabled turns the redirect on. Because the zero value of a bool is
	// false, New defaults Enabled to true when Options are omitted entirely;
	// pass Options{Enabled: false} to explicitly disable.
	Enabled bool
}

// New returns middleware that redirects http requests to https with a 301
// (Moved Permanently). Secure requests, and all requests when disabled, pass
// through untouched. The request is considered secure when it arrived over TLS
// or carries X-Forwarded-Proto: https (both handled by req.Secure).
func New(opts ...Options) express.Handler {
	o := Options{Enabled: true}
	if len(opts) > 0 {
		o = opts[0]
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if !o.Enabled || req.Secure() {
			next()
			return
		}
		res.Redirect(http.StatusMovedPermanently, "https://"+req.Raw.Host+req.OriginalURL())
	}
}
