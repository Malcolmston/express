// Package rawbody provides express middleware that reads the entire request
// body into memory, stores it on the request via SetBody, and restores
// req.Raw.Body so downstream handlers can read it again.
package rawbody

import (
	"bytes"
	"io"

	"github.com/malcolmston/express"
)

// New returns middleware that buffers the full request body into a []byte and
// makes it available through req.Body(). The underlying req.Raw.Body is
// replaced with a fresh reader over the same bytes so subsequent reads (by
// other middleware or handlers) still succeed.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		if req.Raw.Body == nil {
			req.SetBody([]byte{})
			next()
			return
		}
		data, err := io.ReadAll(req.Raw.Body)
		if err != nil {
			next(err)
			return
		}
		req.Raw.Body.Close()
		// Store the raw bytes and restore a re-readable body for downstream.
		req.SetBody(data)
		req.Raw.Body = io.NopCloser(bytes.NewReader(data))
		next()
	}
}
