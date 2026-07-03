// Package bodylimit provides express middleware that rejects requests whose
// body exceeds a configured size, guarding handlers against oversized uploads.
package bodylimit

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the bodylimit middleware.
type Options struct {
	// MaxBytes is the maximum allowed request body size in bytes. A value of
	// 0 or less disables the limit.
	MaxBytes int64
}

// DefaultMaxBytes is the limit applied when New is called with no options.
const DefaultMaxBytes int64 = 1 << 20 // 1 MiB

// New returns middleware enforcing a maximum request body size. If the
// declared Content-Length exceeds the limit the request is rejected
// immediately with 413 Payload Too Large. Otherwise req.Raw.Body is wrapped in
// an http.MaxBytesReader so that a body larger than the limit fails when read.
func New(opts ...Options) express.Handler {
	max := DefaultMaxBytes
	if len(opts) > 0 && opts[0].MaxBytes > 0 {
		max = opts[0].MaxBytes
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		if max <= 0 {
			next()
			return
		}
		if req.Raw.ContentLength > max {
			res.Status(http.StatusRequestEntityTooLarge).
				Type("text").
				Send(http.StatusText(http.StatusRequestEntityTooLarge))
			return
		}
		if req.Raw.Body != nil {
			req.Raw.Body = http.MaxBytesReader(res.Writer, req.Raw.Body, max)
		}
		next()
	}
}
