// Package requestcontext provides middleware that attaches a small
// request-scoped context to each request: a unique request ID and a start
// timestamp. The ID is also echoed in the X-Request-Id response header, easing
// log correlation across services.
package requestcontext

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/malcolmston/express"
)

// Key is the request value key under which the *Ctx is stored.
const Key = "ctx"

// HeaderName is the response header used to expose the request ID.
const HeaderName = "X-Request-Id"

// Ctx is the per-request context attached by this middleware.
type Ctx struct {
	// ID uniquely identifies the request.
	ID string

	// Start is the time the request entered the middleware.
	Start time.Time
}

// Elapsed returns the duration since the request started.
func (c *Ctx) Elapsed() time.Duration { return time.Since(c.Start) }

// Options configures the requestcontext middleware.
type Options struct {
	// TrustHeader, when true, reuses an incoming X-Request-Id header (if
	// present and non-empty) instead of generating a new ID. This is useful
	// behind a trusted proxy that assigns request IDs.
	TrustHeader bool

	// Generator overrides the default random ID generator.
	Generator func() string
}

// New returns middleware that attaches a *Ctx via req.Set(Key, ctx) and sets
// the X-Request-Id response header.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	gen := o.Generator
	if gen == nil {
		gen = generateID
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		id := ""
		if o.TrustHeader {
			id = req.Get(HeaderName)
		}
		if id == "" {
			id = gen()
		}
		ctx := &Ctx{ID: id, Start: time.Now()}
		req.Set(Key, ctx)
		res.Set(HeaderName, id)
		next()
	}
}

// From returns the *Ctx attached to the request, or nil if none is present.
func From(req *express.Request) *Ctx {
	if v, ok := req.Value(Key); ok {
		if c, ok := v.(*Ctx); ok {
			return c
		}
	}
	return nil
}

func generateID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fall back to a timestamp-derived ID; collisions are unlikely and
		// this path is essentially never taken.
		return "req-" + time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b[:])
}
