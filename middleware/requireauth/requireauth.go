// Package requireauth provides middleware that rejects requests unless an
// authenticated principal has already been placed on the request by an earlier
// middleware. It is the express framework's Go analogue of the route guards
// found in the Node ecosystem: connect-ensure-login's ensureLoggedIn(),
// Passport's req.isAuthenticated() check, and the "is this request logged in?"
// gate that most Express apps hand-roll in front of protected routes. Rather
// than performing authentication itself, it enforces that some upstream step
// succeeded and short-circuits everything that did not.
//
// Reach for this middleware to protect a route, a router, or an entire subtree
// once a real authentication middleware runs ahead of it. The typical layout is
// a session, JWT, bearer-token, or API-key middleware that, on success, stashes
// the resolved user on the request (for example req.Set("user", u)); requireauth
// then guarantees that any handler mounted after it only ever runs for an
// authenticated caller. Keeping the "who are you" and "you must be someone"
// concerns in separate middleware lets you mix and match authentication schemes
// while sharing a single, uniform rejection policy.
//
// Operationally the middleware belongs after the authentication step and before
// the protected handlers. On each request it looks up a single request value by
// name using req.Value(Options.Key). When that value is present and non-nil it
// calls next() and the request proceeds unchanged; the middleware neither reads
// nor writes any headers and does not modify the stored value. When the value is
// absent, or present but nil, the request is stopped: the middleware writes a
// 401 Unauthorized status with the plain-text body "Unauthorized" and next() is
// never invoked, so no downstream handler runs.
//
// The only knob is Options.Key, the request-value name to require; it defaults
// to "user" when left empty, matching the de-facto req.user convention. The
// presence test is deliberate about nil: a key explicitly set to a nil value
// (a typed nil interface) counts as unauthenticated and is rejected, so upstream
// code should only set the key once it holds a real principal. Any non-nil value
// of any type satisfies the check — the middleware inspects existence, not shape,
// so it works with strings, integers, structs, or pointers alike. Because every
// failure collapses to the same bare 401, callers cannot distinguish "never
// authenticated" from "authenticated as nil".
//
// Compared with its Node inspirations this port is intentionally minimal. It has
// no notion of sessions, login pages, or redirect-to-login flows (Express's
// ensureLoggedIn redirects browsers to a returnTo URL), it does not read
// req.isAuthenticated() or any framework-specific session object, and it always
// answers with a 401 and a fixed body rather than a configurable status,
// message, or JSON payload. It is purely a presence guard over the request
// value bag: pair it with whatever authentication middleware populates that bag.
package requireauth

import "github.com/malcolmston/express"

// Options configures the require-auth middleware.
type Options struct {
	// Key is the request value name that must be present for the request to
	// proceed. Defaults to "user".
	Key string
}

// New returns middleware that responds with 401 unless the configured request
// value has been set (for example by an authentication middleware).
func New(opts Options) express.Handler {
	key := opts.Key
	if key == "" {
		key = "user"
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		if v, ok := req.Value(key); ok && v != nil {
			next()
			return
		}
		res.Status(401).Send("Unauthorized")
	}
}
