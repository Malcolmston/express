// Package poweredby provides express middleware that sets the X-Powered-By
// response header to a configurable value.
package poweredby

import "github.com/malcolmston/express"

// DefaultValue is used when no Value option is supplied.
const DefaultValue = "Express"

// Options configures the middleware.
type Options struct {
	// Value is written as the X-Powered-By header.
	Value string
}

// New returns middleware that sets the X-Powered-By response header. This
// overrides any framework default so branding can be customized (or hidden by
// supplying an empty-but-explicit value, in which case the header is removed).
func New(opts ...Options) express.Handler {
	value := DefaultValue
	if len(opts) > 0 {
		value = opts[0].Value
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		if value == "" {
			res.Writer.Header().Del("X-Powered-By")
		} else {
			res.Set("X-Powered-By", value)
		}
		next()
	}
}
