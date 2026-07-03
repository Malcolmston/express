// Package querylimit provides middleware that rejects requests whose raw query
// string exceeds a configured maximum length. Oversized requests receive a 414
// URI Too Long response.
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
