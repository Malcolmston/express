// Package ienoopen provides middleware that sets X-Download-Options: noopen,
// preventing Internet Explorer from executing downloads in the site's context.
package ienoopen

import "github.com/malcolmston/express"

// New returns middleware that sets X-Download-Options: noopen.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Download-Options", "noopen")
		next()
	}
}
