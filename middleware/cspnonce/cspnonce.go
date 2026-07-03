// Package cspnonce provides middleware that generates a per-request nonce and
// emits a Content-Security-Policy header whose script-src directive includes
// that nonce, enabling inline scripts to be allow-listed safely.
package cspnonce

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/malcolmston/express"
)

// ContextKey is the key under which the generated nonce is stored on the
// request and in res.Locals.
const ContextKey = "nonce"

// Options configures the CSP nonce middleware.
type Options struct {
	// DefaultSrc sets the default-src directive value (default "'self'").
	DefaultSrc string
	// ScriptSrc holds additional script-src sources placed alongside the
	// generated nonce (default "'self'").
	ScriptSrc string
	// Bytes is the number of random bytes used to build the nonce (default 16).
	Bytes int
}

func (o *Options) applyDefaults() {
	if o.DefaultSrc == "" {
		o.DefaultSrc = "'self'"
	}
	if o.ScriptSrc == "" {
		o.ScriptSrc = "'self'"
	}
	if o.Bytes <= 0 {
		o.Bytes = 16
	}
}

// New returns middleware that generates a nonce, exposes it via
// req.Set("nonce", n) and res.Locals["nonce"], and sets a
// Content-Security-Policy header of the form:
//
//	default-src <DefaultSrc>; script-src <ScriptSrc> 'nonce-<n>'
//
// Retrieve the nonce with Nonce.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	o.applyDefaults()

	return func(req *express.Request, res *express.Response, next express.Next) {
		n := generate(o.Bytes)
		req.Set(ContextKey, n)
		res.Locals[ContextKey] = n

		policy := fmt.Sprintf("default-src %s; script-src %s 'nonce-%s'", o.DefaultSrc, o.ScriptSrc, n)
		res.Set("Content-Security-Policy", policy)
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

func generate(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}
