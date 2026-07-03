// Package etag provides express middleware that computes a strong ETag for the
// response body and answers conditional requests with 304 Not Modified.
package etag

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/malcolmston/express"
)

// captureWriter buffers the response body so an ETag can be computed over the
// complete payload before anything is sent to the client.
type captureWriter struct {
	http.ResponseWriter
	buf         bytes.Buffer
	status      int
	wroteHeader bool
}

func (w *captureWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
}

func (w *captureWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.buf.Write(p)
}

// New returns middleware that sets a strong ETag header (the SHA-1 of the
// response body, hex-encoded and quoted) on every response with a body. When
// the request's If-None-Match matches the computed tag, a 304 Not Modified is
// sent with an empty body; otherwise the buffered body is written unchanged.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		orig := res.Writer
		cw := &captureWriter{ResponseWriter: orig}
		res.Writer = cw
		next()
		res.Writer = orig

		status := cw.status
		if !cw.wroteHeader {
			status = http.StatusOK
		}
		body := cw.buf.Bytes()

		// Nothing to tag: write through as-is.
		if len(body) == 0 {
			orig.WriteHeader(status)
			return
		}

		sum := sha1.Sum(body)
		tag := `"` + hex.EncodeToString(sum[:]) + `"`
		orig.Header().Set("ETag", tag)

		if ifNoneMatch(req.Get("If-None-Match"), tag) {
			// Conditional GET hit: strip body-specific headers and reply 304.
			orig.Header().Del("Content-Type")
			orig.Header().Del("Content-Length")
			orig.WriteHeader(http.StatusNotModified)
			return
		}

		orig.WriteHeader(status)
		_, _ = orig.Write(body)
	}
}

// ifNoneMatch reports whether the If-None-Match header matches tag. A value of
// "*" matches anything; otherwise the comma-separated list is scanned, ignoring
// any weak-validator prefix.
func ifNoneMatch(header, tag string) bool {
	header = strings.TrimSpace(header)
	if header == "" {
		return false
	}
	if header == "*" {
		return true
	}
	for _, part := range strings.Split(header, ",") {
		candidate := strings.TrimSpace(part)
		candidate = strings.TrimPrefix(candidate, "W/")
		if candidate == tag {
			return true
		}
	}
	return false
}
