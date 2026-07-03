// Package hidepoweredby provides middleware that removes the X-Powered-By
// response header (or replaces it with a decoy value) so the server does not
// advertise the technology stack it runs on.
package hidepoweredby

import "github.com/malcolmston/express"

// Options configures the hidepoweredby middleware. The zero value is usable and
// removes the X-Powered-By header entirely.
type Options struct {
	// SetTo, when non-empty, replaces X-Powered-By with this decoy value
	// instead of removing the header.
	SetTo string
}

// New returns middleware that hides or spoofs the X-Powered-By header. The
// header is adjusted via a before-write hook so it takes effect regardless of
// when downstream handlers commit the response.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.OnBeforeWrite(func() {
			if o.SetTo != "" {
				res.Writer.Header().Set("X-Powered-By", o.SetTo)
			} else {
				res.Writer.Header().Del("X-Powered-By")
			}
		})
		next()
	}
}
