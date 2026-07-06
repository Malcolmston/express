// Package basicauth provides HTTP Basic authentication middleware for the
// express framework. It challenges unauthenticated requests with a
// WWW-Authenticate header and rejects invalid credentials with 401. It is the
// Go analogue of Node middleware such as express-basic-auth or the classic
// basic-auth helper, packaged as a drop-in express.Handler.
//
// Use this middleware to protect an entire application or a subtree of routes
// behind a username/password gate without standing up sessions, cookies, or a
// login page. Because the browser's built-in credential dialog is triggered by
// the challenge, Basic auth is a good fit for internal tools, staging
// environments, health dashboards, and machine-to-machine calls where a full
// authentication flow would be overkill. Mount it with app.Use for a global
// guard, or attach it to a specific router or path prefix to protect only part
// of the tree.
//
// Operationally the middleware sits at the front of the chain. On each request
// it reads the Authorization request header, expects a "Basic <base64>" value,
// and decodes it into a username and password separated by the first colon.
// Those credentials are handed to the caller-supplied Options.Verify callback,
// which is the single source of truth for whether access is granted. When
// Verify returns true the middleware calls next() and the request proceeds
// untouched; the credentials are not stored on the request, so downstream
// handlers that need the identity should capture it inside Verify.
//
// When the header is missing, malformed, not Basic, un-decodable, or Verify
// returns false (or is nil), the request is short-circuited: the middleware
// sets a "WWW-Authenticate: Basic realm=..." response header and writes a 401
// Unauthorized body, and next() is never invoked. The realm advertised in the
// challenge comes from Options.Realm and defaults to "Restricted" when empty.
// All failure modes are treated identically and yield the same 401 so that a
// caller cannot distinguish "no header" from "wrong password".
//
// Security note: Basic authentication transmits credentials in every request,
// protected only by base64 encoding, so it must always be layered over TLS.
// The comparison performed inside Verify is entirely the caller's
// responsibility; to resist timing attacks, compare secrets with
// crypto/subtle.ConstantTimeCompare rather than the == operator. Compared with
// the Node originals, this port keeps the same challenge-and-reject semantics
// but is deliberately minimal: it does not ship its own user store, does not
// support the "challenge: false" silent mode, and delegates every credential
// decision to Verify.
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
