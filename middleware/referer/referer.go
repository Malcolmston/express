// Package referer provides middleware that captures the Referer request header,
// parses out its host, and stores the result on the request under the key
// "referer" for downstream handlers such as analytics and anti-CSRF checks. It
// ports the small but recurring Node pattern of reading req.headers.referer (or
// req.get('referrer')), normalizing it, and stashing it on the request/context
// so later handlers need not re-parse the header themselves.
//
// Use it when several downstream consumers care about where a request came
// from: request logging and referral analytics, soft origin checks, or feeding
// a stricter allowlist gate (see the sibling referercheck package). Doing the
// parse once, up front, means handlers read a typed Referer value instead of
// repeatedly pulling and url.Parse-ing a raw header, and it centralizes the
// quirk that the header is misspelled "Referer" in the HTTP standard while some
// clients and codebases use the correct "Referrer".
//
// Mechanically the middleware reads the "Referer" header, falling back to
// "Referrer" when the former is absent, wraps the raw value in a Referer struct,
// and — when the value is non-empty — parses it with net/url to fill in the
// Host field. It then stores the struct via req.Set(Key, ref) and always calls
// next(); it never writes a response, sets no response headers, and never
// short-circuits, so it is purely additive request state. Register it early via
// app.Use so the captured value is available to everything that follows,
// including any gate that decides whether to reject the request.
//
// The captured value is always present but may be zero. A missing header yields
// a Referer with empty URL and empty Host; a present but unparseable header
// yields the raw URL with an empty Host, since a url.Parse error is swallowed
// rather than surfaced. Host is taken from the parsed URL's Host field, which
// includes the port when the referer carries one (for example "example.com:8443")
// — callers that want the bare hostname should strip the port themselves.
// Retrieve the value with From, which returns (Referer, false) when the
// middleware did not run or the stored value is of an unexpected type.
//
// Parity with the Node original is behavioral: like the typical hand-rolled
// middleware it consults both header spellings, tolerates a missing or malformed
// value without erroring, and exposes a parsed host for convenience. It
// deliberately does not make any allow/deny decision — that policy belongs to a
// separate check so that capture and enforcement stay composable — and it does
// not attempt to reconstruct an origin (scheme + host) beyond what url.Parse
// reports.
package referer

import (
	"net/url"

	"github.com/malcolmston/express"
)

// Key is the request value key under which the Referer info is stored.
const Key = "referer"

// Referer holds the raw Referer header and its parsed host.
type Referer struct {
	// URL is the raw Referer header value (may be empty).
	URL string

	// Host is the host component parsed from URL, or "" when absent or
	// unparseable.
	Host string
}

// New returns middleware that stores a Referer via req.Set(Key, ref). Both the
// standard "Referer" and the (rare) correct spelling "Referrer" are consulted.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		raw := req.Get("Referer")
		if raw == "" {
			raw = req.Get("Referrer")
		}
		ref := Referer{URL: raw}
		if raw != "" {
			if u, err := url.Parse(raw); err == nil {
				ref.Host = u.Host
			}
		}
		req.Set(Key, ref)
		next()
	}
}

// From retrieves the Referer stored by this middleware.
func From(req *express.Request) (Referer, bool) {
	v, ok := req.Value(Key)
	if !ok {
		return Referer{}, false
	}
	r, ok := v.(Referer)
	return r, ok
}
