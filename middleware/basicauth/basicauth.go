// Package basicauth provides HTTP Basic authentication middleware for the
// express framework. It challenges unauthenticated requests with a
// WWW-Authenticate header and rejects invalid credentials with 401.
package basicauth

import (
	"encoding/base64"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the Basic authentication middleware.
type Options struct {
	// Realm is the protection space presented to the client in the
	// WWW-Authenticate challenge. Defaults to "Restricted".
	Realm string
	// Verify reports whether the supplied username and password are valid.
	// It is required.
	Verify func(user, pass string) bool
}

// New returns middleware that enforces HTTP Basic authentication. Requests
// without valid credentials receive a 401 response carrying a
// WWW-Authenticate challenge.
func New(opts Options) express.Handler {
	realm := opts.Realm
	if realm == "" {
		realm = "Restricted"
	}
	challenge := `Basic realm="` + realm + `"`
	return func(req *express.Request, res *express.Response, next express.Next) {
		user, pass, ok := parseBasic(req.Get("Authorization"))
		if !ok || opts.Verify == nil || !opts.Verify(user, pass) {
			res.Set("WWW-Authenticate", challenge)
			res.Status(401).Send("Unauthorized")
			return
		}
		next()
	}
}

// parseBasic decodes an "Authorization: Basic <base64>" header into its
// username and password components.
func parseBasic(header string) (user, pass string, ok bool) {
	const prefix = "Basic "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(header[len(prefix):])
	if err != nil {
		return "", "", false
	}
	creds := string(decoded)
	i := strings.IndexByte(creds, ':')
	if i < 0 {
		return "", "", false
	}
	return creds[:i], creds[i+1:], true
}
