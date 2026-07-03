// Package nosniff provides middleware that sets X-Content-Type-Options: nosniff
// to prevent browsers from MIME-sniffing a response away from the declared
// Content-Type.
package nosniff

import "github.com/malcolmston/express"

// New returns middleware that sets X-Content-Type-Options: nosniff.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Content-Type-Options", "nosniff")
		next()
	}
}
