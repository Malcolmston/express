// Package correlationid provides express middleware that assigns a correlation
// id to every request so that a single logical operation can be traced as it
// fans out across multiple services and log streams. It is the Go analogue of
// Node middleware such as express-correlation-id, the correlation-id package,
// and express-request-id, packaged as a drop-in express.Handler that reads and
// writes the X-Correlation-Id header.
//
// Reach for this middleware whenever you need distributed tracing without a
// heavyweight tracing stack: correlating the log lines a request produces in an
// API gateway, an application, and its downstream dependencies, or stitching
// together the story of one client action across a request/response boundary.
// Because an inbound id is preserved rather than replaced, the same value flows
// from the caller through every hop that forwards the header, giving you a
// stable key to group logs, metrics, and error reports by. Mount it globally
// with app.Use so that every route is covered, ideally as one of the first
// handlers in the chain.
//
// Operationally the middleware sits at the very front of the chain and never
// short-circuits. On each request it reads the configured header (Header,
// defaulting to DefaultHeader, "X-Correlation-Id"). If the caller already
// supplied a value, that value is kept; otherwise a new id is produced by the
// Generator callback. The chosen id is then stored on the request via
// req.Set(ContextKey, id) — where ContextKey is "correlationId" — so downstream
// handlers can retrieve it with req.Value(ContextKey), and it is echoed back on
// the response under the same header. The middleware always calls next(), so
// the request proceeds unmodified apart from the stored value and the response
// header.
//
// The two configuration knobs live on Options and both have defaults, so New()
// with no arguments is fully usable. Header selects the header name to read and
// write. Generator produces the id when the incoming request has none; when it
// is nil the built-in generator returns a random 16-byte value encoded as a
// 32-character hex string, falling back to an all-zero string only if the
// system random source fails. There is no failure or rejection path: a missing
// or malformed inbound id is simply treated as "absent" and a fresh id is
// generated, so the middleware can never block a request. Note that the id is
// trusted verbatim when present, so if inbound values must conform to a format
// (for example, to keep logs tidy or resist log injection), validate or
// normalize them inside a custom Generator-aware pipeline before relying on
// them.
//
// Compared with the Node originals, this port keeps the essential
// preserve-or-generate contract and the X-Correlation-Id convention, but is
// deliberately minimal. It does not integrate with continuation-local storage
// (as cls-rtracer does), does not expose a global "current id" accessor outside
// the request, and stores the id on the express Request rather than in a
// package-level async context. Everything a handler needs is available through
// req.Value(ContextKey) and the response header.
package correlationid

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/malcolmston/express"
)

// DefaultHeader is the header used when none is configured.
const DefaultHeader = "X-Correlation-Id"

// ContextKey is the key under which the id is stored via req.Set.
const ContextKey = "correlationId"

// Options configures the middleware.
type Options struct {
	// Header is the request/response header carrying the correlation id.
	// Defaults to X-Correlation-Id.
	Header string
	// Generator produces a new id when the incoming request has none. When
	// nil, a random 16-byte hex id is generated.
	Generator func() string
}

// New returns middleware that ensures every request carries a correlation id.
// An incoming id in the configured header is preserved so the same value flows
// through downstream service calls; otherwise a new id is generated. The id is
// echoed on the response and stored on the request via req.Set(ContextKey, id).
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
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b)
}
