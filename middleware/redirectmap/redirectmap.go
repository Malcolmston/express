// Package redirectmap provides middleware that redirects requests whose path
// matches an entry in a static lookup table. It ports the pattern of a static
// redirect map — the kind of "old path -> new path" table often expressed as a
// plain object handed to Node's express-urlrewrite, a hand-written redirect
// middleware, or the redirects section of a hosting provider's config — into a
// single express handler backed by a Go map.
//
// Use it to migrate or consolidate URLs without registering an individual route
// for each one: renamed pages, retired endpoints that should point at their
// replacement, vanity links, or a bulk import of legacy paths. Because lookups
// are a single map access, a table with thousands of entries costs the same as
// a table with one, which makes it a convenient place to park large,
// data-driven redirect sets that would be tedious to wire up as routes.
//
// Mechanically the handler looks up req.Path() in the table. On a hit it calls
// res.Redirect(status, dest) and returns without invoking next(), so the
// response is fully written (a Location header plus the redirect status) and the
// chain is short-circuited. On a miss it calls next() and touches nothing else,
// letting the request flow on to routers and other middleware, which the
// FallThrough test confirms. It reads only the request path and writes only the
// Location header and status line, so it is safe to register early via app.Use.
//
// Behavior is controlled by two options. Map holds the path-to-destination
// entries and is matched by exact path equality — there is no prefix, glob, or
// query-string matching, and the request query string is ignored during lookup.
// Status is the redirect code and defaults to 302 (Found) when left zero; set it
// to 301 (Moved Permanently) for permanent moves that clients and search engines
// should cache, as the CustomStatus test demonstrates. The provided map is
// copied at construction time, so mutating the caller's map afterward cannot
// alter the middleware's behavior, and an empty or nil map simply falls every
// request through.
//
// Parity with the Node originals is intentionally minimal and predictable.
// Where express-urlrewrite and similar packages support regular-expression
// patterns, capture groups, and internal rewrites (changing the URL without a
// client round-trip), this package restricts itself to exact-path external
// redirects with a configurable status. That narrower contract keeps lookups
// O(1) and the semantics obvious, and callers who need pattern matching can
// layer a dedicated rewrite router in front of it.
package redirectmap

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the redirect-map middleware.
type Options struct {
	// Map associates request paths with destination URLs. A request whose
	// path is a key is redirected to the corresponding value.
	Map map[string]string

	// Status is the HTTP status code used for redirects. When zero it
	// defaults to 302 (Found).
	Status int
}

// New returns middleware that redirects any request whose path is present in
// the map to the mapped destination; other requests fall through to next.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	status := o.Status
	if status == 0 {
		status = http.StatusFound
	}
	// Copy the map so later mutation by the caller cannot affect behavior.
	m := make(map[string]string, len(o.Map))
	for k, v := range o.Map {
		m[k] = v
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if dest, ok := m[req.Path()]; ok {
			res.Redirect(status, dest)
			return
		}
		next()
	}
}
