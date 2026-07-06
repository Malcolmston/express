// Package frameguard provides middleware that sets the X-Frame-Options response
// header to control whether the page may be rendered inside a <frame>,
// <iframe>, <embed>, or <object>, helping to mitigate clickjacking attacks. It
// is a stdlib-only port of Helmet's "frameguard" module (historically the
// standalone "frameguard" npm package), which is one of the header-setting
// middlewares Helmet enables by default.
//
// Clickjacking tricks a user into interacting with your site while it is framed
// invisibly beneath attacker-controlled UI. Setting X-Frame-Options tells the
// browser to refuse to render the response inside a frame from another origin,
// defeating the attack for legacy browsers. Use this middleware on any HTML
// endpoint that does not need to be embedded cross-origin. Modern browsers also
// honor the frame-ancestors directive of Content-Security-Policy, which
// supersedes X-Frame-Options; frameguard remains valuable for older clients and
// as defense in depth, and the two can be sent together.
//
// The middleware sits anywhere before the response is written, though mounting
// it globally with app.Use near the top of the chain is typical so the header
// is applied to every response. On each request it calls res.Set to write the
// X-Frame-Options header to the chosen directive and then calls next() to
// continue the chain. It never reads request state, never inspects or writes the
// body, and never short-circuits, so it composes freely with other middleware
// and route handlers.
//
// Behavior is governed by the single Action option. The zero value of Options
// is usable and yields the default directive "SAMEORIGIN", which permits
// framing only by pages on the same origin. The other supported value is "DENY",
// which forbids all framing regardless of origin; matching is case-insensitive,
// so "deny" is normalized to "DENY". Any other non-empty value — including the
// legacy "ALLOW-FROM uri" form, which this port does not implement — falls back
// to "SAMEORIGIN" rather than emitting an invalid or unsupported directive.
//
// Compared with Helmet's frameguard the port covers the two directives that
// current browsers still respect (DENY and SAMEORIGIN) and deliberately omits
// the deprecated ALLOW-FROM option, which most browsers ignore and which was
// removed from Helmet itself. The header is always set (there is no way to
// conditionally skip it per request), so if a particular route must be
// embeddable cross-origin you should not mount frameguard on that route, or you
// should override the header downstream.
package frameguard

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the frameguard middleware. The zero value is usable and
// yields X-Frame-Options: SAMEORIGIN.
type Options struct {
	// Action is the X-Frame-Options directive to send. Supported values are
	// "SAMEORIGIN" (default) and "DENY" (case-insensitive).
	Action string
}

// New returns middleware that sets the X-Frame-Options header.
func New(opts ...Options) express.Handler {
	action := "SAMEORIGIN"
	if len(opts) > 0 && opts[0].Action != "" {
		if strings.EqualFold(opts[0].Action, "DENY") {
			action = "DENY"
		} else {
			action = "SAMEORIGIN"
		}
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Frame-Options", action)
		next()
	}
}
