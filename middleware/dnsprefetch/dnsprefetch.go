// Package dnsprefetch provides middleware that sets the X-DNS-Prefetch-Control
// response header, controlling whether browsers speculatively resolve the DNS
// names of links and resources on a page. It is the Go analogue of Helmet's
// dnsPrefetchControl middleware from the Node ecosystem, packaged as a drop-in
// express.Handler that writes the same header with the same "on"/"off" semantics.
//
// Use this middleware to make a deliberate choice about DNS prefetching rather
// than leaving it to browser defaults. Prefetching improves perceived
// performance by resolving hostnames before the user clicks, but it also leaks a
// signal to DNS resolvers about which links a page contains, which can be a
// minor privacy concern for pages that reference sensitive third-party hosts.
// Turning it off (the default here) prioritizes privacy; turning it on
// prioritizes latency. Mount it with app.Use for a site-wide policy, or attach
// it to a specific router when only some pages warrant a different setting.
//
// Operationally the middleware belongs anywhere before the response body is
// written. On each request it sets a single response header,
// X-DNS-Prefetch-Control, to a value computed once when New builds the handler,
// and then always calls next() so the request proceeds untouched. It reads no
// request headers and no request state; it is purely additive to the response
// and never short-circuits the chain.
//
// The behavior is governed by Options.Allow. When Allow is true the header value
// is "on", enabling prefetching; when false — including the zero-value Options,
// which is usable — the value is "off", disabling it. New is variadic and
// accepts zero or one Options; calling New() with no arguments is equivalent to
// New(Options{}) and yields "off". There is no failure mode and no per-request
// branching beyond writing the precomputed value.
//
// Compared with Helmet's dnsPrefetchControl, this port keeps the identical
// contract — one header, "on" or "off", defaulting to "off" — but is
// intentionally minimal: it exposes a single boolean rather than an options
// object, and it does not attempt to emit any related headers. It pairs well
// with the sibling security-header middlewares (such as csp and dnsprefetch's
// Helmet cousins) when composing a broader hardening layer.
package dnsprefetch

import "github.com/malcolmston/express"

// Options configures the dnsprefetch middleware. The zero value is usable and
// yields X-DNS-Prefetch-Control: off.
type Options struct {
	// Allow enables DNS prefetching ("on"). When false, prefetching is turned
	// "off".
	Allow bool
}

// New returns middleware that sets the X-DNS-Prefetch-Control header.
func New(opts ...Options) express.Handler {
	value := "off"
	if len(opts) > 0 && opts[0].Allow {
		value = "on"
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-DNS-Prefetch-Control", value)
		next()
	}
}
