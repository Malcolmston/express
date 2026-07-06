// Package querylimit provides middleware that rejects requests whose raw query
// string exceeds a configured maximum length. Oversized requests receive a 414
// URI Too Long response before any route handler runs.
//
// It plays the same defensive role as guards found around Node's query parsers
// (for example the parameterLimit / maximum URL length caps in qs and
// body-parser, and the URL-length limits enforced by reverse proxies): it
// bounds the amount of query-string data the application will accept so that a
// single request cannot force expensive parsing of an enormous or maliciously
// crafted query. Rather than counting individual parameters, this port caps the
// total byte length of the raw query string, which is a cheap, allocation-free
// check performed before the query is parsed.
//
// Use it when endpoints accept user-supplied query parameters and you want a
// hard upper bound on their size to protect against denial-of-service or memory
// pressure from abusive URLs. Register it early with app.Use, ahead of any
// middleware that parses or iterates the query (such as querynormalize or route
// handlers that call req.Query), so oversized requests are rejected before that
// work happens.
//
// On each request the middleware reads req.Raw.URL.RawQuery (the portion after
// the leading '?', not including it) and compares its byte length against
// MaxLength. When the length is within the limit it calls next() and the chain
// proceeds normally; no headers or state are modified. When the length exceeds
// the limit it short-circuits: it sets the status to 414 (http.StatusRequest
// URITooLong) via res.Status, writes Message as the body via res.Send, and
// returns without calling next(), so downstream handlers never run.
//
// Options control the threshold and the rejection body. MaxLength is measured in
// bytes and values <= 0 fall back to a default of 2048; Message defaults to
// "URI Too Long" when empty. Note the comparison is strict (len > MaxLength), so
// a query exactly MaxLength bytes long is accepted. The check operates on the
// raw, still-encoded query string, meaning percent-encoded sequences count by
// their encoded byte length.
//
// There is no single Express package this mirrors one-to-one; stock Express
// delegates query parsing to qs and relies on the underlying HTTP server and
// any fronting proxy for URL-length limits. This middleware makes that limit an
// explicit, in-application, configurable policy with a conventional 414 response.
package querylimit

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the query-limit middleware.
type Options struct {
	// MaxLength is the maximum permitted length, in bytes, of the raw query
	// string (excluding the leading '?'). Values <= 0 default to 2048.
	MaxLength int
	// Message is the response body sent on rejection. Defaults to
	// "URI Too Long".
	Message string
}

// New returns query-limit middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.MaxLength <= 0 {
		o.MaxLength = 2048
	}
	if o.Message == "" {
		o.Message = "URI Too Long"
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if len(req.Raw.URL.RawQuery) > o.MaxLength {
			res.Status(http.StatusRequestURITooLong).Send(o.Message)
			return
		}
		next()
	}
}
