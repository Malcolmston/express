// Package crossoriginresource provides middleware that sets the
// Cross-Origin-Resource-Policy (CORP) response header, which limits which
// origins may include the resource.
package crossoriginresource

import "github.com/malcolmston/express"

// Options configures the crossoriginresource middleware. The zero value is
// usable and yields Cross-Origin-Resource-Policy: same-origin.
type Options struct {
	// Policy overrides the header value (e.g. "same-origin", "same-site",
	// "cross-origin"). When empty, "same-origin" is used.
	Policy string
}

// New returns middleware that sets the Cross-Origin-Resource-Policy header.
func New(opts ...Options) express.Handler {
	value := "same-origin"
	if len(opts) > 0 && opts[0].Policy != "" {
		value = opts[0].Policy
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Cross-Origin-Resource-Policy", value)
		next()
	}
}
