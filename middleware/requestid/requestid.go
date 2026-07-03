// Package requestid provides express middleware that assigns each request a
// unique identifier, echoes it on the response, and stores it on the request
// for downstream handlers and logging.
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
