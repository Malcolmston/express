// Package hmacauth provides middleware that authenticates requests by
// verifying an HMAC-SHA256 signature of the request body against a shared
// secret.
package hmacauth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/malcolmston/express"
)

// Options configures the HMAC authentication middleware.
type Options struct {
	// Secret is the shared HMAC key. Required.
	Secret []byte
	// Header names the request header carrying the hex-encoded signature.
	// Defaults to "X-Signature".
	Header string
}

// New returns middleware that computes the HMAC-SHA256 of the raw request body
// and compares it, in constant time, to the hex value supplied in the
// configured header. Requests failing verification are rejected with 401. The
// request body is buffered and restored so downstream handlers can still read
// it.
func New(opts Options) express.Handler {
	header := opts.Header
	if header == "" {
		header = "X-Signature"
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		var body []byte
		if req.Raw.Body != nil {
			b, err := io.ReadAll(req.Raw.Body)
			if err != nil {
				res.Status(401).Send("Unauthorized")
				return
			}
			body = b
			req.Raw.Body.Close()
			// Restore the body for downstream handlers.
			req.Raw.Body = io.NopCloser(bytes.NewReader(body))
		}

		mac := hmac.New(sha256.New, opts.Secret)
		mac.Write(body)
		expected := mac.Sum(nil)

		provided, err := hex.DecodeString(req.Get(header))
		if err != nil || !hmac.Equal(expected, provided) {
			res.Status(401).Send("Unauthorized")
			return
		}
		next()
	}
}

// Sign returns the hex-encoded HMAC-SHA256 of body using secret. It is useful
// for clients and tests constructing signed requests.
func Sign(secret, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
