// Package hmacauth provides middleware that authenticates requests by
// verifying an HMAC-SHA256 signature of the request body against a shared
// secret. It fills the role of Node webhook-verification helpers such as the
// signature checks used by GitHub, Stripe, and Shopify webhooks, or the generic
// "express-hmac-auth" style middleware, implemented here for this Express-style
// framework using only crypto/hmac and crypto/sha256 from the standard library.
//
// Use it to protect endpoints — most commonly inbound webhooks and
// service-to-service APIs — where both peers share a secret key and the caller
// signs each request. Verifying an HMAC over the body proves two things at
// once: that the sender knew the secret (authentication) and that the body was
// not altered in transit (integrity). Unlike a static bearer token, the
// signature changes with every payload, so a captured signature cannot be
// replayed against a different body.
//
// In the chain the middleware must run before any handler that consumes the
// request body, typically as an app.Use guarding a route group. On each request
// it reads the raw request body in full via req.Raw.Body, computes the
// HMAC-SHA256 of those exact bytes with the configured Secret, and compares it
// in constant time (hmac.Equal) against the hex-decoded value taken from the
// configured request header (req.Get). Because reading the body consumes it,
// the middleware buffers the bytes and restores req.Raw.Body with a fresh
// io.NopCloser so downstream handlers can read the body normally. On success it
// calls next(); on any failure it short-circuits with 401 Unauthorized and does
// not call next().
//
// A request is rejected with 401 in every failure mode: the body cannot be
// read, the header is missing or not valid hex, or the decoded signature does
// not equal the computed one (including a signature that was valid for a
// different body). Options.Secret is required — an empty secret still produces a
// deterministic MAC but offers no security, so always supply a strong random
// key. Options.Header defaults to "X-Signature". The constant-time comparison
// avoids leaking timing information, but note the entire body is buffered in
// memory, so pair this with a body-size limit for untrusted clients. The
// exported Sign helper computes the same hex signature clients must send.
//
// Parity note: this port fixes the algorithm to HMAC-SHA256 over the raw body
// and a hex-encoded signature header. It does not implement the timestamp
// tolerance, replay-nonce tracking, or configurable digest algorithms that some
// provider-specific verifiers add; if you need those, layer them around this
// middleware. The signing scheme is intentionally simple and symmetric so that
// Sign and New always agree.
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
