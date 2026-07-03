// Package apikey provides middleware that authenticates requests using an API
// key supplied via a request header or query-string parameter.
package apikey

import (
	"crypto/subtle"

	"github.com/malcolmston/express"
)

// Options configures the API-key middleware.
type Options struct {
	// Header names the request header carrying the key. Defaults to
	// "X-API-Key".
	Header string
	// Query, when non-empty, additionally accepts the key from this
	// query-string parameter.
	Query string
	// Keys is the set of accepted keys. Ignored when Verify is set.
	Keys []string
	// Verify, when set, takes precedence over Keys and reports whether a key
	// is valid.
	Verify func(key string) bool
}

// New returns middleware that requires a valid API key. Missing or invalid
// keys are rejected with 401.
func New(opts Options) express.Handler {
	header := opts.Header
	if header == "" {
		header = "X-API-Key"
	}
	verify := opts.Verify
	if verify == nil {
		keys := opts.Keys
		verify = func(key string) bool {
			for _, k := range keys {
				if subtle.ConstantTimeCompare([]byte(k), []byte(key)) == 1 {
					return true
				}
			}
			return false
		}
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		key := req.Get(header)
		if key == "" && opts.Query != "" {
			key = req.Query(opts.Query)
		}
		if key == "" || !verify(key) {
			res.Status(401).Send("Unauthorized")
			return
		}
		next()
	}
}
