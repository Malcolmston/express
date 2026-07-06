// Package apikey provides middleware that authenticates requests using an API
// key supplied via a request header or query-string parameter. It is the Go
// analogue of the many Node API-key gatekeepers (for example
// express-api-key-auth style middleware) and complements the package's
// basicauth middleware, packaged as a drop-in express.Handler that rejects
// unauthenticated requests with 401.
//
// Use this middleware to protect an application or a subtree of routes behind a
// shared secret without sessions, cookies, or a login page. It fits internal
// services, server-to-server calls, and public APIs that issue per-client
// keys, where a lightweight bearer-style credential is more appropriate than
// interactive authentication. Mount it with app.Use for a global guard or
// attach it to a specific router to protect only part of the tree.
//
// Operationally the middleware sits at the front of the chain. On each request
// it reads the key from the configured header (Options.Header, default
// "X-API-Key") via req.Get; if that is empty and Options.Query is set, it
// falls back to reading the key from that query-string parameter via
// req.Query. The extracted key is then validated. On success next() is called
// and the request proceeds untouched; the key is not stored on the request, so
// a handler that needs the caller's identity should derive it inside a Verify
// callback.
//
// Validation is governed by two mutually exclusive options. When
// Options.Verify is set it is the sole authority and Options.Keys is ignored,
// letting you look keys up in a database or apply custom rules. When Verify is
// nil the middleware compares the presented key against each entry in
// Options.Keys using crypto/subtle.ConstantTimeCompare, so the built-in check
// resists timing attacks. A missing key (empty from both header and query) or
// one that fails validation short-circuits with res.Status(401).Send, and
// next() is never invoked; every failure yields the same 401 so a caller
// cannot distinguish "no key" from "wrong key".
//
// Security note: like all bearer credentials, an API key is exposed in every
// request and must be sent over TLS; passing it in the query string is
// convenient but risks leaking the key into server and proxy logs, so prefer
// the header when possible. Compared with the Node originals this port is
// deliberately minimal: it ships no key store, rotation, scoping, or rate
// limiting, and delegates any such policy to the Verify callback while
// providing a safe constant-time default for static key lists.
package apikey

import (
	"crypto/subtle"

	"github.com/malcolmston/express"
)

// Options configures the API-key middleware.
type Options struct {
	// Header names the request header carrying the key. Defaults to
	// "X-API-Key".
	Header string
	// Query, when non-empty, additionally accepts the key from this
	// query-string parameter.
	Query string
	// Keys is the set of accepted keys. Ignored when Verify is set.
	Keys []string
	// Verify, when set, takes precedence over Keys and reports whether a key
	// is valid.
	Verify func(key string) bool
}

// New returns middleware that requires a valid API key. Missing or invalid
// keys are rejected with 401.
func New(opts Options) express.Handler {
	header := opts.Header
	if header == "" {
		header = "X-API-Key"
	}
	verify := opts.Verify
	if verify == nil {
		keys := opts.Keys
		verify = func(key string) bool {
			for _, k := range keys {
				if subtle.ConstantTimeCompare([]byte(k), []byte(key)) == 1 {
					return true
				}
			}
			return false
		}
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		key := req.Get(header)
		if key == "" && opts.Query != "" {
			key = req.Query(opts.Query)
		}
		if key == "" || !verify(key) {
			res.Status(401).Send("Unauthorized")
			return
		}
		next()
	}
}
