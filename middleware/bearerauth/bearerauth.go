// Package bearerauth provides middleware that authenticates requests using an
// opaque bearer token supplied in the Authorization header.
package bearerauth

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the bearer authentication middleware.
type Options struct {
	// Verify reports whether the presented token is valid. It is required.
	Verify func(token string) bool
}

// New returns middleware that reads an "Authorization: Bearer <token>" header
// and rejects the request with 401 when the token is missing or invalid.
func New(opts Options) express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		token, ok := extractToken(req.Get("Authorization"))
		if !ok || opts.Verify == nil || !opts.Verify(token) {
			res.Set("WWW-Authenticate", "Bearer")
			res.Status(401).Send("Unauthorized")
			return
		}
		next()
	}
}

// extractToken pulls the token out of a "Bearer <token>" authorization header.
func extractToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}
