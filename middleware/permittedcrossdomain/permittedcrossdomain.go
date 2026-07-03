// Package permittedcrossdomain provides middleware that sets the
// X-Permitted-Cross-Domain-Policies response header, controlling Adobe
// Flash/Acrobat cross-domain policy files.
package permittedcrossdomain

import "github.com/malcolmston/express"

// Options configures the permittedcrossdomain middleware. The zero value is
// usable and yields X-Permitted-Cross-Domain-Policies: none.
type Options struct {
	// Policy overrides the header value (e.g. "none", "master-only",
	// "by-content-type", "all"). When empty, "none" is used.
	Policy string
}

// New returns middleware that sets the X-Permitted-Cross-Domain-Policies header.
func New(opts ...Options) express.Handler {
	value := "none"
	if len(opts) > 0 && opts[0].Policy != "" {
		value = opts[0].Policy
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Permitted-Cross-Domain-Policies", value)
		next()
	}
}
