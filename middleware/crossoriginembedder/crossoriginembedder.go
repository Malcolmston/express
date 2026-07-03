// Package crossoriginembedder provides middleware that sets the
// Cross-Origin-Embedder-Policy (COEP) response header, which controls whether a
// document may load cross-origin resources that lack explicit permission.
package crossoriginembedder

import "github.com/malcolmston/express"

// Options configures the crossoriginembedder middleware. The zero value is
// usable and yields Cross-Origin-Embedder-Policy: require-corp.
type Options struct {
	// Policy overrides the header value (e.g. "require-corp",
	// "credentialless", "unsafe-none"). When empty, "require-corp" is used.
	Policy string
}

// New returns middleware that sets the Cross-Origin-Embedder-Policy header.
func New(opts ...Options) express.Handler {
	value := "require-corp"
	if len(opts) > 0 && opts[0].Policy != "" {
		value = opts[0].Policy
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Cross-Origin-Embedder-Policy", value)
		next()
	}
}
