// Package cookieparser provides express middleware that parses every cookie
// from the incoming request into a map[string]string and stores it on the
// request for convenient downstream access. It is the Go analogue of the Node
// cookie-parser middleware (expressjs/cookie-parser), packaged as a drop-in
// express.Handler, and it turns the raw Cookie request header into a lookup map
// so handlers need not re-parse it themselves.
//
// Use this middleware whenever downstream handlers or other middleware want to
// read cookies by name — session identifiers, feature flags, CSRF tokens,
// preferences, and the like. Mount it once near the top of the chain with
// app.Use so the parsed map is available everywhere, or attach it to a specific
// router or path prefix if only part of the tree needs cookie access. It only
// reads request cookies; setting cookies on the response remains the job of
// res.Cookie.
//
// Operationally the middleware runs early and does all its work before calling
// next(). On each request it iterates req.Raw.Cookies(), and for each cookie it
// attempts to URL-unescape the value with url.QueryUnescape. When unescaping
// succeeds the decoded value is stored; when it fails the raw, still-encoded
// value is stored instead, so a malformed cookie never drops out of the map.
// The resulting map is attached to the request via req.Set under the key
// "cookies", and then next() is invoked. The middleware never inspects headers
// beyond the incoming cookies, sets nothing on the response, and never
// short-circuits the chain.
//
// Retrieve the parsed cookies with the From helper, which reads the stored map
// back off the request. From is defensive: if the middleware never ran, or the
// stored value is somehow not a map, it returns an empty but non-nil
// map[string]string, so callers can range over or index the result without a
// nil check. When a cookie name appears more than once, the last occurrence in
// req.Raw.Cookies() order wins, since each write to the map overwrites the
// previous value for that key.
//
// Compared with the Node original, this port keeps the same "parse once, read
// by name" convenience but is deliberately minimal. It does not support signed
// cookies or a secret (see the signedcookies or cookiesession middleware for
// tamper protection), it does not perform JSON cookie decoding, and it exposes
// values as plain strings in a flat map rather than distinguishing signed and
// unsigned collections. Unescaping uses URL query decoding rather than Node's
// cookie-specific decoder, which covers the common encodeURIComponent case.
package cookieparser

import (
	"net/url"

	"github.com/malcolmston/express"
)

// contextKey is the key under which the parsed cookie map is stored on the
// request via req.Set.
const contextKey = "cookies"

// New returns middleware that parses every cookie on the request into a
// map[string]string and stores it via req.Set("cookies", m). Cookie values are
// URL-unescaped when possible. Retrieve the map with From.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		m := make(map[string]string)
		for _, c := range req.Raw.Cookies() {
			if v, err := url.QueryUnescape(c.Value); err == nil {
				m[c.Name] = v
			} else {
				m[c.Name] = c.Value
			}
		}
		req.Set(contextKey, m)
		next()
	}
}

// From returns the parsed cookie map previously stored by the middleware. If
// the middleware did not run it returns an empty, non-nil map.
func From(req *express.Request) map[string]string {
	if v, ok := req.Value(contextKey); ok {
		if m, ok := v.(map[string]string); ok {
			return m
		}
	}
	return map[string]string{}
}
