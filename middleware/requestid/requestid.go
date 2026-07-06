// Package requestid provides express middleware that assigns each request a
// unique identifier, echoes it on the response, and stores it on the request
// for downstream handlers and logging. It is the express port of the Node
// express-request-id middleware and the widely used X-Request-ID convention:
// every request is guaranteed to carry an ID that ties together application
// logs, client-visible response headers, and, when propagated, traces spanning
// multiple services.
//
// Use it as a foundational layer for observability. With a request ID in place,
// a log line, an error report, and the response the client received can all be
// correlated by the same value, which turns "a user saw an error at 3pm" into a
// searchable key. It is also the natural hook for distributed tracing at the
// edge: an upstream proxy or gateway can stamp X-Request-Id and this middleware
// will honor it, so the same ID flows through your service instead of a new one
// being minted at each hop.
//
// Mechanically the middleware runs early in the chain and, for each request,
// reads the configured header from the incoming request; if that value is empty
// it generates a fresh ID; then it stores the ID on the request via
// req.Set(ContextKey, id) and mirrors it onto the response with
// res.Set(header, id) before calling next() to continue. It never
// short-circuits and leaves the status and body to downstream handlers.
// Registering it with app.Use ahead of any logger or handler that reads
// req.Value(ContextKey) is required, and placing it early ensures the response
// header is present even on error responses.
//
// Behavior is governed by Options with sensible defaults. Options.Header selects
// the request/response header and defaults to DefaultHeader ("X-Request-Id");
// the ID is stored under the constant ContextKey ("requestId"). Options.Generator
// overrides ID minting and, when nil, a cryptographically random 16-byte value
// rendered as a 32-character hex string is used; in the essentially-unreachable
// case that crypto/rand fails, a fixed all-zero 32-character ID is returned
// rather than panicking. Reuse of an inbound ID is unconditional: any non-empty
// value in the configured header is trusted and echoed verbatim, so if clients
// are untrusted you should strip or override the header upstream, or supply a
// Generator that ignores inbound values.
//
// Parity with the Node original is behavioral: like express-request-id this
// package guarantees an ID on every request, prefers an existing header value,
// exposes the ID on the response, and makes it retrievable by downstream code.
// It keeps the surface minimal — a configurable header and generator — rather
// than reproducing every option of the JavaScript package, and it uses Go's
// crypto/rand hex format as the default ID shape.
package requestid

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/malcolmston/express"
)

// DefaultHeader is the header used when none is configured.
const DefaultHeader = "X-Request-Id"

// ContextKey is the key under which the id is stored via req.Set.
const ContextKey = "requestId"

// Options configures the middleware.
type Options struct {
	// Header is the request/response header carrying the id. Defaults to
	// X-Request-Id.
	Header string
	// Generator produces a new id when the incoming request has none. When
	// nil, a random 16-byte hex id is generated.
	Generator func() string
}

// New returns middleware that ensures every request carries an id. If the
// incoming request already has one in the configured header it is reused;
// otherwise a fresh id is generated. The id is echoed on the response header
// and stored on the request via req.Set(ContextKey, id).
func New(opts ...Options) express.Handler {
	header := DefaultHeader
	gen := generateID
	if len(opts) > 0 {
		if opts[0].Header != "" {
			header = opts[0].Header
		}
		if opts[0].Generator != nil {
			gen = opts[0].Generator
		}
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		id := req.Get(header)
		if id == "" {
			id = gen()
		}
		req.Set(ContextKey, id)
		res.Set(header, id)
		next()
	}
}

// generateID returns a random 16-byte value as a 32-character hex string.
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// rand.Read on crypto/rand effectively never fails; fall back to a
		// fixed-length zero id rather than panicking.
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b)
}
