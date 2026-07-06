// Package bodylimit provides express middleware that rejects requests whose
// body exceeds a configured size, guarding handlers against oversized uploads.
// It is the Go analogue of the size limit enforced by Node's body-parser and
// the raw-body module (the "limit" option surfaced by express.json,
// express.urlencoded, and express.raw), extracted into a standalone
// express.Handler that can protect any route regardless of content type.
//
// Use this middleware to defend an application against memory-exhaustion and
// denial-of-service attempts in which a client streams an unbounded request
// body. Because it caps the number of bytes any downstream handler can read, it
// is valuable in front of file-upload endpoints, JSON APIs, and proxy routes,
// and is cheap enough to mount globally with app.Use so that every route
// inherits a sane ceiling. Set a smaller per-route limit on endpoints that
// should only ever receive tiny payloads.
//
// The middleware sits at the front of the chain and enforces the limit in two
// complementary ways. First, if the request's declared Content-Length header
// already exceeds the limit, the request is rejected immediately with 413
// Payload Too Large and next() is never called, so an oversized upload is
// refused before a single byte of the body is read. Otherwise it wraps
// req.Raw.Body in an http.MaxBytesReader keyed to res.Writer and calls next();
// a handler that then tries to read more than the limit receives an error from
// io.Read, and the underlying server records the overflow. This second gate
// catches bodies sent with a missing, unknown, or dishonest Content-Length.
//
// Configuration is a single field, Options.MaxBytes, giving the maximum body
// size in bytes. New is variadic: calling it with no options applies
// DefaultMaxBytes (1 MiB), while a MaxBytes value of 0 or less disables the
// limit entirely and lets every request pass through untouched. Only the first
// Options value is consulted. Because the check is expressed purely in bytes it
// makes no attempt to interpret the payload; charset, encoding, and
// content-type concerns are left to whatever parser runs later in the chain.
//
// A caveat worth noting versus the Node original: body-parser measures the
// decoded body length and can account for content encoding, whereas this
// middleware limits the raw wire bytes as delivered. The 413 responses it
// writes carry a plain-text body and, unlike body-parser, it does not
// distinguish "entity too large" from "request aborted" — both oversized
// declarations and oversized reads are treated as the same limit violation.
package bodylimit

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the bodylimit middleware.
type Options struct {
	// MaxBytes is the maximum allowed request body size in bytes. A value of
	// 0 or less disables the limit.
	MaxBytes int64
}

// DefaultMaxBytes is the limit applied when New is called with no options.
const DefaultMaxBytes int64 = 1 << 20 // 1 MiB

// New returns middleware enforcing a maximum request body size. If the
// declared Content-Length exceeds the limit the request is rejected
// immediately with 413 Payload Too Large. Otherwise req.Raw.Body is wrapped in
// an http.MaxBytesReader so that a body larger than the limit fails when read.
func New(opts ...Options) express.Handler {
	max := DefaultMaxBytes
	if len(opts) > 0 && opts[0].MaxBytes > 0 {
		max = opts[0].MaxBytes
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		if max <= 0 {
			next()
			return
		}
		if req.Raw.ContentLength > max {
			res.Status(http.StatusRequestEntityTooLarge).
				Type("text").
				Send(http.StatusText(http.StatusRequestEntityTooLarge))
			return
		}
		if req.Raw.Body != nil {
			req.Raw.Body = http.MaxBytesReader(res.Writer, req.Raw.Body, max)
		}
		next()
	}
}
