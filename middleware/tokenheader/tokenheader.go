// Package tokenheader provides generic middleware that validates a token
// supplied in a single configurable request header.
package tokenheader

import "github.com/malcolmston/express"

// Options configures the token-header middleware.
type Options struct {
	// Header names the request header carrying the token. Required.
	Header string
	// Verify reports whether the presented token is valid. Required.
	Verify func(token string) bool
}

// New returns middleware that reads the configured header and rejects the
// request with 401 when the token is missing or fails verification.
func New(opts Options) express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		token := ""
		if opts.Header != "" {
			token = req.Get(opts.Header)
		}
		if token == "" || opts.Verify == nil || !opts.Verify(token) {
			res.Status(401).Send("Unauthorized")
			return
		}
		next()
	}
}
