// Package expires provides express middleware that sets an HTTP Expires header a
// fixed duration into the future. It fills the same role as the Node
// "express-expires" style helpers and the Expires option of static-file
// middleware such as serve-static: a relative freshness lifetime expressed as a
// duration is turned into an absolute HTTP-date on every response.
//
// Use this middleware to give responses a coarse cache lifetime for HTTP/1.0
// caches and proxies that predate Cache-Control, or simply as a belt-and-braces
// companion to Cache-Control: max-age. It is well suited to content that should
// be considered fresh for a fixed window after it is served — semi-static
// pages, infrequently changing JSON, downloadable assets. Mount it with app.Use
// over the routes whose responses should carry an expiry.
//
// The handler runs anywhere in the chain and writes exactly one header. On each
// request it computes now().Add(Duration).UTC(), formats it with
// http.TimeFormat (the RFC 1123 GMT form required for HTTP dates), calls
// res.Set("Expires", value), and then invokes next() unconditionally. It never
// reads the request, never short-circuits, and does not touch the status or
// body; the header is set before next() so it is present when a downstream
// handler flushes the response. Unlike a header value computed once at
// construction time, the Expires timestamp is evaluated per request, so it
// always reflects the moment each request is handled plus the configured offset.
//
// Behavior is governed by Options.Duration, which is added to the current time.
// A zero Duration (the zero-value Options) produces an Expires equal to the
// current time — effectively "expires immediately" — while a negative Duration
// yields a date in the past, marking the response already stale. The package
// uses an internal, overridable now function (defaulting to time.Now) so tests
// can pin the clock; production callers never interact with it. No upper bound
// is imposed on the duration, and the value is always emitted in UTC/GMT as the
// HTTP-date spec requires.
//
// Regarding parity with the Node original: the semantics match the common
// Express pattern of translating a relative max-age-like duration into an
// absolute Expires header formatted as an HTTP date. The differences are that
// this port sets only the Expires header and does not also emit Cache-Control
// (callers who want both should pair it with a cache-control middleware), and
// that it takes a Go time.Duration rather than a milliseconds number or a date
// string. It performs no conditional-request or freshness checking; it only
// stamps the header.
package expires

import (
	"net/http"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the expires middleware.
type Options struct {
	// Duration is added to the current time to produce the Expires value.
	Duration time.Duration
}

// now is overridable in tests.
var now = time.Now

// New returns middleware that sets the Expires response header to the current
// time plus Duration, formatted as an HTTP date (RFC 1123 GMT).
func New(opts ...Options) express.Handler {
	var d time.Duration
	if len(opts) > 0 {
		d = opts[0].Duration
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		t := now().Add(d).UTC()
		res.Set("Expires", t.Format(http.TimeFormat))
		next()
	}
}
