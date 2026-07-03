// Package contenttypedefault provides express middleware that sets a default
// Content-Type on the response when a handler does not set one itself.
package contenttypedefault

import "github.com/malcolmston/express"

// DefaultType is used when no Type option is supplied.
const DefaultType = "application/octet-stream"

// Options configures the middleware.
type Options struct {
	// Type is the Content-Type applied when none is present at write time.
	Type string
}

// New returns middleware that, just before the response headers are committed,
// sets Content-Type to the configured default if the handler has not already
// set one.
func New(opts ...Options) express.Handler {
	typ := DefaultType
	if len(opts) > 0 && opts[0].Type != "" {
		typ = opts[0].Type
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.OnBeforeWrite(func() {
			if res.GetHeader("Content-Type") == "" {
				res.Set("Content-Type", typ)
			}
		})
		next()
	}
}
