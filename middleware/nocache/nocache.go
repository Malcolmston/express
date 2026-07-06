// Package nocache provides express middleware that instructs clients and
// proxies never to cache the response. It is the express port of the Node
// nocache package (the same helper popularized by helmet's legacy noCache
// module and by lusca), which sets the classic trio of anti-caching headers
// that HTTP/1.0 proxies, HTTP/1.1 caches, and older browsers each pay attention
// to.
//
// Reach for it on any endpoint whose response must always be revalidated:
// authenticated pages, dynamic dashboards, CSRF-token or nonce carrying HTML,
// account and billing views, and API responses that reflect frequently changing
// or user-specific state. Serving such content with default caching can leak one
// user's data to another through a shared proxy or let a stale, sensitive page
// linger in the browser's back/forward cache, so making "do not store this"
// explicit is both a correctness and a security measure.
//
// The middleware runs early in the chain and writes three response headers
// before calling next() so that downstream handlers inherit them: Cache-Control
// is set to "no-store, no-cache, must-revalidate" to forbid storage and force
// revalidation, Pragma is set to "no-cache" for HTTP/1.0 intermediaries, and
// Expires is set to "0" so any date-based cache treats the response as already
// stale. It reads no request state, never short-circuits, and always continues
// the chain by invoking next().
//
// A few semantics are worth noting. There are no options: the header set is
// fixed, matching the conservative defaults of the Node original. Because the
// headers are written before next() runs, a later handler is free to override
// them (for example to opt a specific sub-response back into caching), and the
// last write wins. Note that this package does not set an ETag or Last-Modified
// and does not strip them if a prior handler added them; "must-revalidate" plus
// "no-cache" is what disciplines any conditional requests. The header values are
// deliberately belt-and-suspenders so that the response is uncacheable across the
// full range of HTTP/1.0 and HTTP/1.1 caching behaviors.
//
// Parity with the Node nocache middleware is faithful: it emits the same
// Cache-Control/Pragma/Expires combination and, like the original, leaves the
// Surrogate-Control CDN header alone since Go deployments typically manage edge
// caching separately. If you need per-route cache directives instead of a blanket
// no-store policy, prefer the cachecontrol middleware; nocache is the blunt,
// always-off instrument.
package nocache

import "github.com/malcolmston/express"

// New returns middleware that sets the standard set of no-cache headers:
// Cache-Control, Pragma, and Expires. It is useful for dynamic, sensitive, or
// frequently changing responses.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Cache-Control", "no-store, no-cache, must-revalidate")
		res.Set("Pragma", "no-cache")
		res.Set("Expires", "0")
		next()
	}
}
