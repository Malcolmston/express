// Package methodallow provides middleware that restricts requests to a
// configured set of HTTP methods, responding with 405 Method Not Allowed
// otherwise. It is the Go equivalent of the small method-guard helpers used in
// the Node/Express world (for example connect/express snippets or packages such
// as method-not-allowed) that reject unexpected verbs and advertise the allowed
// set through an Allow header, implemented here with only the standard library.
//
// Use this middleware when a route or an entire application should accept only a
// known subset of HTTP methods and you want unsupported verbs answered with the
// correct 405 status and an Allow header instead of falling through to a generic
// 404 or an unintended handler. It is a good fit for read-only endpoints
// (GET/HEAD only), write endpoints that should reject everything but POST, or
// any surface where you want to make the permitted verbs explicit and
// discoverable to clients and API tooling.
//
// Register it ahead of the handlers it should guard, typically via app.Use so
// it fronts the routes, or scoped to a sub-router. At construction time New
// upper-cases every entry in Options.Methods into a lookup set and precomputes
// the Allow header value (the methods joined with ", " in the order given). On
// each request it upper-cases req.Method() and checks membership: if the method
// is permitted it calls next() and the request proceeds; if not it sets the
// Allow header and writes status 405 with the body "Method Not Allowed",
// returning without calling next() so no downstream handler runs.
//
// Method matching is case-insensitive because both the configured list and the
// incoming method are canonicalized to upper case, so listing "get" or "GET" is
// equivalent. Options.Methods is required and has no default: constructing the
// middleware with an empty slice yields an empty allow set and an empty Allow
// header, which rejects every request with 405 — configure the intended verbs
// explicitly. The middleware does not special-case HEAD or OPTIONS; if you want
// those to pass you must include them in Methods.
//
// Security-wise, returning 405 with an accurate Allow header is the
// standards-correct way to refuse an unsupported verb and helps prevent verb
// tampering from reaching handlers that never expected it. Compared with the
// Node originals, this port keeps the essential contract — allow a fixed set,
// otherwise 405 plus Allow — while using Go idioms: a precomputed map and header
// string built once in New, canonical upper-case comparison, and a plain-text
// body rather than the configurable error objects some JavaScript variants emit.
package methodallow

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the method-allow middleware.
type Options struct {
	// Methods lists the permitted HTTP methods (case-insensitive). Required.
	Methods []string
}

// New returns middleware that responds with 405 Method Not Allowed, including
// an Allow header, unless the request method is in the permitted set.
func New(opts Options) express.Handler {
	allowed := make(map[string]struct{}, len(opts.Methods))
	canonical := make([]string, 0, len(opts.Methods))
	for _, m := range opts.Methods {
		up := strings.ToUpper(m)
		allowed[up] = struct{}{}
		canonical = append(canonical, up)
	}
	allowHeader := strings.Join(canonical, ", ")
	return func(req *express.Request, res *express.Response, next express.Next) {
		if _, ok := allowed[strings.ToUpper(req.Method())]; ok {
			next()
			return
		}
		res.Set("Allow", allowHeader)
		res.Status(405).Send("Method Not Allowed")
	}
}
