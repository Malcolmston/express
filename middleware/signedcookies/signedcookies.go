// Package signedcookies provides middleware that verifies a tamper-evident,
// HMAC-signed cookie and exposes its plaintext value to downstream handlers.
// It is the express framework's Go analogue of the signed-cookie support built
// into Express via the cookie-parser middleware (expressjs/cookie-parser) and
// the underlying cookie-signature module (tj/node-cookie-signature): a value is
// stored on the client alongside a keyed HMAC so the server can detect any
// modification without keeping server-side session state.
//
// Reach for this middleware when you place a small, non-secret identifier in a
// cookie — a user id, a tenant slug, a feature flag — and need to trust it on
// the way back in without a database round trip. The signature guarantees
// integrity and authenticity (only a holder of Secret could have produced it),
// so a client cannot forge or edit the value. It does not provide
// confidentiality: the plaintext value is visible in the cookie, so never place
// a password or other secret in it; encrypt separately if the payload must be
// hidden.
//
// Operationally the middleware belongs near the front of the chain, before any
// handler that trusts the cookie's value. On each request it reads the cookie
// named Options.Name via req.Cookie. When present it splits the stored string on
// the last '.' into value and hex signature, recomputes the HMAC-SHA256 of the
// value under Secret, and compares it with hmac.Equal (a constant-time compare
// that resists timing attacks). On a match it stashes the verified value on the
// request with req.Set(Key, value) — where Key defaults to Name — and calls
// next(); a downstream handler retrieves it with req.Value(Key). The middleware
// never writes the cookie itself; use the exported Sign helper to mint the
// cookie value at login and set it with res.Cookie.
//
// Semantics and edge cases center on the two short-circuit paths. A missing
// cookie normally stops the request with a 401 Unauthorized and next() is never
// called; set Options.Optional to true to instead let anonymous requests
// proceed (nothing is stored, so req.Value(Key) reports absent). A present but
// tampered, truncated, or unsigned cookie — including any value with no '.'
// separator — always fails verification and yields the same 401. Secret and
// Name are required; supplying a nil or wrong Secret makes every previously
// issued cookie fail to verify, which is the intended behavior for key
// rotation.
//
// Compared with the Node originals this port is deliberately minimal. It reads
// exactly one named cookie rather than sweeping every cookie into a
// req.signedCookies map, it uses HMAC-SHA256 with a raw hex encoding instead of
// cookie-signature's base64 SHA-256 "value.signature" scheme (so signatures are
// not wire-compatible with cookie-parser), and it performs the rejection itself
// rather than merely marking a bad cookie as false. It ships no cookie writer,
// no expiry or rotation of multiple secrets, and no encryption; those concerns
// are left to the caller and to res.Cookie.
package signedcookies

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the signed-cookie middleware.
type Options struct {
	// Secret is the HMAC key used to sign and verify cookie values. Required.
	Secret []byte
	// Name is the cookie to read. Required.
	Name string
	// Key is the request value name under which the verified cookie value is
	// stored. Defaults to the cookie Name.
	Key string
	// Optional, when true, allows requests without the cookie to proceed
	// instead of being rejected with 401.
	Optional bool
}

// Sign returns a signed cookie value of the form "value.hexHMAC" produced from
// value and secret.
func Sign(secret []byte, value string) string {
	return value + "." + mac(secret, value)
}

// Verify checks a signed cookie value and returns the embedded value when the
// signature is valid.
func Verify(secret []byte, signed string) (string, bool) {
	i := strings.LastIndexByte(signed, '.')
	if i < 0 {
		return "", false
	}
	value, sig := signed[:i], signed[i+1:]
	expected := mac(secret, value)
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return "", false
	}
	return value, true
}

// New returns middleware that verifies the configured signed cookie. On
// success the underlying value is stored on the request; on failure the
// request is rejected with 401 unless Optional is set.
func New(opts Options) express.Handler {
	key := opts.Key
	if key == "" {
		key = opts.Name
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		raw := req.Cookie(opts.Name)
		if raw == "" {
			if opts.Optional {
				next()
				return
			}
			res.Status(401).Send("Unauthorized")
			return
		}
		value, ok := Verify(opts.Secret, raw)
		if !ok {
			res.Status(401).Send("Unauthorized")
			return
		}
		req.Set(key, value)
		next()
	}
}

// mac returns the hex-encoded HMAC-SHA256 of value under secret.
func mac(secret []byte, value string) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}
