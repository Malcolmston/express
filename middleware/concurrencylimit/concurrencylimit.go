// Package concurrencylimit provides middleware that caps the number of
// requests processed concurrently. It uses a buffered channel as a counting
// semaphore. When the limit is reached, additional requests are rejected
// immediately with a 503 Service Unavailable response rather than queueing.
package concurrencylimit

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the concurrency limiter.
type Options struct {
	// Max is the maximum number of in-flight requests. Values <= 0 default to 1.
	Max int
	// Message is the body sent when the limit is exceeded. Defaults to
	// "Service Unavailable".
	Message string
}

// New returns concurrency-limiting middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Max <= 0 {
		o.Max = 1
	}
	if o.Message == "" {
		o.Message = "Service Unavailable"
	}

	sem := make(chan struct{}, o.Max)

	return func(req *express.Request, res *express.Response, next express.Next) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			next()
		default:
			res.Status(http.StatusServiceUnavailable).Send(o.Message)
		}
	}
}
