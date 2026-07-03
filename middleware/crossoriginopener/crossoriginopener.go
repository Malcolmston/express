// Package crossoriginopener provides middleware that sets the
// Cross-Origin-Opener-Policy (COOP) response header, which controls whether a
// document may share a browsing context group with cross-origin documents.
package crossoriginopener

import "github.com/malcolmston/express"

// Options configures the crossoriginopener middleware. The zero value is usable
// and yields Cross-Origin-Opener-Policy: same-origin.
type Options struct {
	// Policy overrides the header value (e.g. "same-origin",
	// "same-origin-allow-popups", "unsafe-none"). When empty, "same-origin" is
	// used.
	Policy string
}

// New returns middleware that sets the Cross-Origin-Opener-Policy header.
func New(opts ...Options) express.Handler {
	value := "same-origin"
	if len(opts) > 0 && opts[0].Policy != "" {
		value = opts[0].Policy
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Cross-Origin-Opener-Policy", value)
		next()
	}
}
