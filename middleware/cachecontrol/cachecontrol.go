// Package cachecontrol provides express middleware that sets a Cache-Control
// header assembled from a set of caching options. It plays the role of Node's
// nocache middleware and the cache-control helpers built into serve-static and
// express.static (their maxAge/cacheControl options), but generalized to any
// route so that a single mount can express both "cache this aggressively" and
// "never store this" policies as a reusable express.Handler.
//
// Use this middleware to centralize caching policy instead of scattering
// res.Set("Cache-Control", ...) calls across handlers. Mount it globally with
// app.Use to give a whole application a default policy, or attach it to a
// specific router or path prefix — long max-age with "public" for immutable
// static assets, "no-store" for authenticated or sensitive JSON endpoints, and
// "private" for per-user responses that shared caches must not retain. Because
// the header value is computed once when the middleware is built, per-request
// overhead is a single header write.
//
// Operationally the middleware runs early, before the handler that generates
// the body, so the header is in place regardless of how the response is later
// produced. On each request it writes the precomputed Cache-Control value with
// res.Set and then always calls next(); it never short-circuits the chain and
// never inspects the request. A downstream handler remains free to overwrite or
// delete the header, so this middleware establishes a default rather than an
// immutable policy.
//
// The directive string is assembled from Options in a stable, deterministic
// order — "public", "private", "no-store", then "max-age=N" — and joined with
// ", ". The max-age directive is emitted when MaxAge is greater than zero, or
// when SetMaxAge is true, which is the only way to emit an explicit
// "max-age=0". Note that Public and Private are independent booleans: setting
// both emits a contradictory "public, private" pair, so callers should choose
// one. When every option is left at its zero value the computed string is
// empty and the middleware sets no header at all, passing the request through
// untouched.
//
// Compared with the Node originals this port is intentionally small: it models
// only the handful of directives exposed through Options and does not implement
// no-cache, must-revalidate, s-maxage, immutable, stale-while-revalidate, or
// ETag/Last-Modified negotiation. It also performs no validation of
// conflicting combinations, emitting exactly the directives the options
// request. For anything beyond these directives, set the header directly in a
// handler.
package cachecontrol

import (
	"strconv"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the Cache-Control directive.
type Options struct {
	// MaxAge is the max-age directive in seconds. It is emitted when greater
	// than zero (or when SetMaxAge is true, allowing an explicit max-age=0).
	MaxAge int
	// SetMaxAge forces emission of max-age even when MaxAge is 0.
	SetMaxAge bool
	// Public adds the "public" directive.
	Public bool
	// Private adds the "private" directive.
	Private bool
	// NoStore adds the "no-store" directive.
	NoStore bool
}

// New returns middleware that sets the Cache-Control response header built from
// the provided options. Directives are emitted in a stable order.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	value := build(o)
	return func(req *express.Request, res *express.Response, next express.Next) {
		if value != "" {
			res.Set("Cache-Control", value)
		}
		next()
	}
}

func build(o Options) string {
	var parts []string
	if o.Public {
		parts = append(parts, "public")
	}
	if o.Private {
		parts = append(parts, "private")
	}
	if o.NoStore {
		parts = append(parts, "no-store")
	}
	if o.MaxAge > 0 || o.SetMaxAge {
		parts = append(parts, "max-age="+strconv.Itoa(o.MaxAge))
	}
	return strings.Join(parts, ", ")
}
