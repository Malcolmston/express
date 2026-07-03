// Package requireauth provides middleware that rejects requests unless an
// authenticated user has been placed on the request by earlier middleware.
package requireauth

import "github.com/malcolmston/express"

// Options configures the require-auth middleware.
type Options struct {
	// Key is the request value name that must be present for the request to
	// proceed. Defaults to "user".
	Key string
}

// New returns middleware that responds with 401 unless the configured request
// value has been set (for example by an authentication middleware).
func New(opts Options) express.Handler {
	key := opts.Key
	if key == "" {
		key = "user"
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		if v, ok := req.Value(key); ok && v != nil {
			next()
			return
		}
		res.Status(401).Send("Unauthorized")
	}
}
