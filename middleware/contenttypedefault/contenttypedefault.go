// Package contenttypedefault provides express middleware that sets a fallback
// Content-Type on the response when a downstream handler does not set one
// itself. It is the Go analogue of the small connect/express helper pattern
// that guarantees a default media type (the server-side counterpart of the
// implicit Content-Type Express applies via res.type), packaged as a drop-in
// express.Handler. It exists so responses never go out with a missing or
// accidentally sniffed Content-Type.
//
// Use this middleware when handlers in your application may write raw bytes
// without declaring a media type and you want a single, predictable fallback
// instead of relying on the client to sniff the body. It is especially useful
// for binary or opaque payloads, where the safe default of
// "application/octet-stream" tells the client to download rather than render.
// Mount it globally with app.Use so every response is covered, or attach it to
// a specific router or path prefix to give one subtree its own default.
//
// Operationally the middleware runs early but defers its decision to write
// time. On each request it registers a res.OnBeforeWrite callback and then
// immediately calls next() so the chain proceeds normally. The callback fires
// once, just before the response headers are committed, and only then checks
// whether a Content-Type is already present. This late binding is what lets a
// handler that runs after this middleware still win: whatever media type the
// handler set (via res.Type, res.Set, res.Send, res.JSON, and so on) is
// observed at commit time and left untouched.
//
// Semantically the middleware is purely additive and never overrides. If
// res.GetHeader("Content-Type") is non-empty when the headers are about to be
// written, nothing happens; only a genuinely absent Content-Type is filled in.
// The value applied comes from Options.Type and defaults to DefaultType
// ("application/octet-stream") when Type is empty or no Options are supplied.
// The middleware sets no other headers, changes no status code, and never
// short-circuits the chain.
//
// Compared with the ad-hoc Node snippets it mirrors, this port keeps the same
// "fill only if missing" contract but ties the check to the framework's
// OnBeforeWrite hook rather than monkey-patching the response object, which
// makes the ordering deterministic regardless of where in the chain the
// downstream handler sets its type. It is deliberately tiny: a single Type
// option, one default, and no content negotiation or charset handling of its
// own.
package contenttypedefault

import "github.com/malcolmston/express"

// DefaultType is used when no Type option is supplied.
const DefaultType = "application/octet-stream"

// Options configures the middleware.
type Options struct {
	// Type is the Content-Type applied when none is present at write time.
	Type string
}

// New returns middleware that, just before the response headers are committed,
// sets Content-Type to the configured default if the handler has not already
// set one.
func New(opts ...Options) express.Handler {
	typ := DefaultType
	if len(opts) > 0 && opts[0].Type != "" {
		typ = opts[0].Type
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.OnBeforeWrite(func() {
			if res.GetHeader("Content-Type") == "" {
				res.Set("Content-Type", typ)
			}
		})
		next()
	}
}
