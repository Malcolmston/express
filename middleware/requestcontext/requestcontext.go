// Package requestcontext provides middleware that attaches a small
// request-scoped context to each request: a unique request ID and a start
// timestamp. It is the express analogue of the request-scoped value injection
// commonly done in Node with a context.Context-style object or libraries such
// as express-http-context and cls-hooked, where per-request data is stashed so
// any downstream handler can retrieve it without threading parameters through
// every call. The ID is also echoed in the X-Request-Id response header, easing
// log correlation across services.
//
// Use it when handlers, loggers, or instrumentation deeper in the chain need a
// stable identifier and a request start time without each layer re-deriving
// them. Because the *Ctx is stored on the request itself, a logging helper or
// error formatter can call From(req) and immediately obtain the same ID that
// was returned to the client, which is the foundation of tracing a single
// request across application logs and, when propagated, across service
// boundaries. The Ctx.Elapsed helper turns the captured Start time into a
// running request duration suitable for latency logging.
//
// Mechanically the middleware runs early in the chain and, for each request,
// resolves an ID, records the current time, builds a *Ctx, stores it via
// req.Set(Key, ctx), and mirrors the ID onto the response with
// res.Set(HeaderName, id) before calling next() to continue. It never
// short-circuits: the response body and status are left entirely to downstream
// handlers. Registering it with app.Use before any handler or logger that
// calls From is required, since From returns nil when no context has been
// attached yet.
//
// ID resolution has two modes governed by Options. By default a fresh random
// 16-byte hex ID is generated per request via crypto/rand (falling back to a
// timestamp-derived value in the essentially-unreachable case that rand.Read
// fails). When Options.TrustHeader is true the middleware first reads any
// inbound X-Request-Id header and reuses it when present and non-empty,
// generating a new ID only as a fallback; this is intended for deployment
// behind a trusted proxy or gateway that already assigns IDs. Options.Generator
// overrides the ID source entirely, which is useful for deterministic tests or
// for adopting a different ID format (UUID, ULID, and so on). Note that
// TrustHeader should only be enabled when the header source is trusted, since a
// client-supplied X-Request-Id would otherwise be reflected verbatim.
//
// Parity with the Node original is behavioral rather than API-identical: like
// express-request-id and the ambient-context helpers it targets, this package
// guarantees every request carries an ID, exposes it on the response, and makes
// it retrievable downstream. It deliberately keeps a tiny, typed surface — a
// *Ctx with ID and Start plus the From accessor — instead of reproducing the
// full continuation-local-storage machinery of cls-hooked, which is unnecessary
// given that the value travels on the express Request.
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
