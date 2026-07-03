// Package signedcookies provides middleware that verifies a tamper-evident,
// HMAC-signed cookie and exposes its value to downstream handlers.
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
