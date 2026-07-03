// Package compression provides express middleware that gzip-compresses
// responses for clients that advertise gzip support via Accept-Encoding.
package compression

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/malcolmston/express"
)

// DefaultMinLength is the smallest response body (in bytes) that is compressed
// when no MinLength is configured. Very small payloads are cheaper to send
// uncompressed.
const DefaultMinLength = 256

// Options configures the compression middleware.
type Options struct {
	// Level is the gzip compression level (gzip.BestSpeed ..
	// gzip.BestCompression). Zero selects gzip.DefaultCompression.
	Level int
	// MinLength is the minimum body size, in bytes, eligible for compression.
	// Zero selects DefaultMinLength.
	MinLength int
}

// captureWriter buffers the response body so the middleware can decide whether
// (and how) to compress it once the handler has finished.
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

// New returns middleware that gzip-compresses eligible responses. A response is
// compressed only when the client's Accept-Encoding includes gzip, the body is
// at least MinLength bytes, and no Content-Encoding is already set.
func New(opts ...Options) express.Handler {
	level := gzip.DefaultCompression
	minLength := DefaultMinLength
	if len(opts) > 0 {
		if opts[0].Level != 0 {
			level = opts[0].Level
		}
		if opts[0].MinLength > 0 {
			minLength = opts[0].MinLength
		}
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		if !acceptsGzip(req.Get("Accept-Encoding")) {
			next()
			return
		}

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

		// Fall back to an uncompressed passthrough when compression is not
		// worthwhile or not permitted.
		if len(body) < minLength || orig.Header().Get("Content-Encoding") != "" {
			orig.WriteHeader(status)
			_, _ = orig.Write(body)
			return
		}

		h := orig.Header()
		h.Set("Content-Encoding", "gzip")
		h.Add("Vary", "Accept-Encoding")
		h.Del("Content-Length") // length changes after compression
		orig.WriteHeader(status)

		gw, err := gzip.NewWriterLevel(orig, level)
		if err != nil {
			_, _ = orig.Write(body)
			return
		}
		_, _ = gw.Write(body)
		_ = gw.Close()
	}
}

// acceptsGzip reports whether an Accept-Encoding header allows gzip.
func acceptsGzip(header string) bool {
	for _, part := range strings.Split(header, ",") {
		token := strings.TrimSpace(part)
		if i := strings.IndexByte(token, ';'); i >= 0 {
			token = strings.TrimSpace(token[:i])
		}
		if strings.EqualFold(token, "gzip") {
			return true
		}
	}
	return false
}
