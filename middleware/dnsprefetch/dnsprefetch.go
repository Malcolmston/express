// Package dnsprefetch provides middleware that sets the X-DNS-Prefetch-Control
// response header, controlling browser DNS prefetching.
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
