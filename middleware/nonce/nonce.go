// Package nonce provides middleware that generates a fresh, random,
// base64-encoded nonce for each request. It mirrors the Node helpers built
// around Content-Security-Policy nonces — for example express's res.locals
// nonce pattern and the cspNonce/helmet-csp integration used to authorize
// specific inline scripts and styles — by producing one unguessable token per
// request that downstream code can attach to a CSP header and to markup.
//
// Reach for it whenever you need a per-request nonce, most commonly to allow a
// handful of trusted inline <script>/<style> blocks under a strict CSP without
// resorting to 'unsafe-inline'. Because a nonce is only safe if it is fresh and
// unpredictable for every response, this middleware regenerates it on each
// request rather than reusing a value; templates and a CSP builder then read the
// same token so the header and the markup agree.
//
// It should run early in the chain, before any CSP-setting middleware and before
// your view renders. On each request it calls Generate to produce the token, then
// publishes it in two places so both worlds can see it: req.Set(ContextKey, n)
// for downstream Go handlers (retrievable with Nonce) and res.Locals[ContextKey]
// for templates. It writes no response headers itself and always calls next() to
// continue the chain — wiring the nonce into an actual CSP header is the job of a
// separate middleware or your route.
//
// The token is built with crypto/rand and base64-standard encoded. Options.Bytes
// controls the entropy and defaults to 16 bytes when unset or non-positive, which
// after base64 yields a comfortably unguessable string. ContextKey is the fixed
// lookup key ("nonce") shared by the request store and Locals. One edge case to
// know: if the system random source fails, Generate returns an empty string
// rather than panicking, so a defensive CSP builder should treat "" as "no nonce
// available"; Nonce likewise returns "" when the middleware never ran or stored a
// non-string value.
//
// Parity with the Node original is behavioral rather than byte-for-byte. Like the
// common express nonce recipe it exposes the value through res.Locals and
// generates cryptographically strong randomness per request; the base64 alphabet
// and default byte count match the widely used defaults. It intentionally stops at
// producing and publishing the nonce, leaving header composition to the csp/
// cspnonce middleware so the two concerns stay decoupled.
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
