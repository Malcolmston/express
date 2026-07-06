// Package rawbody provides express middleware that reads the entire request
// body into memory, stores it on the request via SetBody, and restores
// req.Raw.Body so downstream handlers can read it again.
//
// It ports the buffering behaviour of Node's raw-body module (the same building
// block body-parser uses under the hood): it drains the incoming request stream
// once, into a single in-memory []byte, and makes that buffer available to the
// rest of the application. Capturing the raw bytes is essential for tasks that
// must see the exact payload — HMAC/webhook signature verification, request
// logging or auditing, or custom parsing — where re-reading the network stream
// would otherwise fail because an io.Reader can only be consumed once.
//
// Use it when more than one consumer needs the body, or when a handler needs the
// unmodified bytes rather than a decoded value. Register it with app.Use ahead
// of any middleware or handler that inspects the payload (for example a
// signature checker followed by a JSON decoder). Because it buffers eagerly on
// every matching request, pair it with a size guard such as the bodylimit
// middleware in front of it when accepting untrusted input, so a large upload
// cannot exhaust memory.
//
// On each request the middleware reads req.Raw.Body to completion with
// io.ReadAll, closes the original body, stores the resulting bytes on the
// request with req.SetBody (retrievable as a []byte through req.Body), and
// installs a fresh, re-readable body by wrapping the same bytes in an
// io.NopCloser over a bytes.Reader. Downstream code can therefore both call
// req.Body to get the buffer directly and read req.Raw.Body again as a normal
// stream. It writes no response headers.
//
// Edge cases are handled explicitly. When req.Raw.Body is nil (as with a
// typical GET), the middleware stores an empty, non-nil []byte via SetBody and
// calls next() without touching the stream, so req.Body always yields a []byte.
// When io.ReadAll fails, it does not swallow the error: it calls next(err),
// propagating the failure into the framework's error-handling chain (which by
// default produces a 500 response) rather than continuing to the handler. On
// the success path it calls next() with no argument to continue normally.
//
// Compared with the Node original, the scope is narrower: raw-body in Node also
// enforces a byte length limit, honors a declared encoding to optionally return
// a decoded string, and can reject on a Content-Length mismatch. This port keeps
// only the core capture-and-restore behaviour, always yielding raw bytes and
// delegating size limiting to dedicated middleware such as bodylimit.
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
