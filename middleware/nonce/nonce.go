// Package nonce provides middleware that generates a fresh, random,
// base64-encoded nonce for each request. The nonce is exposed both on the
// request and in res.Locals so templates and downstream middleware (for
// example a Content-Security-Policy builder) can reference it.
package nonce

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/malcolmston/express"
)

// ContextKey is the key under which the generated nonce is stored on the
// request and in res.Locals.
const ContextKey = "nonce"

// Options configures the nonce middleware.
type Options struct {
	// Bytes is the number of random bytes used to build the nonce (default 16).
	Bytes int
}

// New returns middleware that stores a per-request random nonce via
// req.Set("nonce", n) and res.Locals["nonce"]. Retrieve it with Nonce.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Bytes <= 0 {
		o.Bytes = 16
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		n := Generate(o.Bytes)
		req.Set(ContextKey, n)
		res.Locals[ContextKey] = n
		next()
	}
}

// Nonce returns the nonce generated for the request, or "" if the middleware
// did not run.
func Nonce(req *express.Request) string {
	if v, ok := req.Value(ContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Generate returns a random base64-encoded nonce built from n random bytes.
func Generate(n int) string {
	if n <= 0 {
		n = 16
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}
