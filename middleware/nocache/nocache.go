// Package nocache provides express middleware that instructs clients and
// proxies never to cache the response.
package nocache

import "github.com/malcolmston/express"

// New returns middleware that sets the standard set of no-cache headers:
// Cache-Control, Pragma, and Expires. It is useful for dynamic, sensitive, or
// frequently changing responses.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Cache-Control", "no-store, no-cache, must-revalidate")
		res.Set("Pragma", "no-cache")
		res.Set("Expires", "0")
		next()
	}
}
