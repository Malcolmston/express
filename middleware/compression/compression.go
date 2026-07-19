// Package compression provides express middleware that gzip-compresses
// response bodies for clients that advertise gzip support via the
// Accept-Encoding request header. It is the Go analogue of the Node
// compression middleware (expressjs/compression), packaged as a drop-in
// express.Handler, and it trades a small amount of CPU for a large reduction
// in bytes on the wire for compressible payloads such as HTML, JSON, CSS, and
// JavaScript.
//
// Use this middleware when you serve text-based responses to bandwidth-limited
// or latency-sensitive clients and want smaller transfers without touching
// your handlers. Mount it once near the top of the chain with app.Use so that
// it wraps every downstream response, or attach it to a specific router or
// path prefix to compress only part of the tree. Because it buffers the body
// to make a size-based decision, it is best suited to ordinary buffered
// responses rather than long-lived streams; place any handler that needs raw,
// unbuffered access to the response writer ahead of it.
//
// Operationally the middleware runs early but does its real work after the
// downstream handler returns. On each request it first inspects
// Accept-Encoding: if the client does not list gzip it calls next() and leaves
// the response completely untouched. Otherwise it swaps res.Writer for an
// internal capturing writer, calls next() so the handler writes as usual, and
// then restores the original writer. The captured status code and body are
// then examined to decide whether compression is worthwhile. When it does
// compress, it sets Content-Encoding: gzip, adds Vary: Accept-Encoding so
// caches key on the negotiated encoding, deletes the now-incorrect
// Content-Length, writes the captured status, and streams the gzipped bytes to
// the original writer.
//
// Several conditions short-circuit compression and cause the buffered body to
// be written through verbatim: a body smaller than MinLength (default
// DefaultMinLength, 256 bytes), or a response that already carries a
// Content-Encoding header (for example an image or a pre-compressed asset).
// The compression Level defaults to gzip.DefaultCompression and may be set to
// any value from gzip.BestSpeed to gzip.BestCompression; an invalid level that
// causes gzip.NewWriterLevel to fail results in the uncompressed body being
// written as a safe fallback. Because the body is buffered in memory before
// being flushed, very large responses hold their full size in memory for the
// duration of the request.
//
// Compared with the Node original, this port keeps the same negotiate,
// buffer, and conditionally-encode model and the same Vary and Content-Length
// handling, but is deliberately narrower in scope. It supports only gzip (not
// deflate or brotli), it does not consult a per-Content-Type filter function
// to decide compressibility, and it does not expose a streaming flush hook;
// the decision to compress rests solely on Accept-Encoding, the body length,
// and any existing Content-Encoding.
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

// WriteHeader implements http.ResponseWriter; it records the first status code
// without writing it to the client so the header can be sent later once the
// buffered body is compressed.
func (w *captureWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
}

// Write implements http.ResponseWriter; it buffers p in memory (defaulting the
// status to 200 on first write) instead of writing to the client immediately.
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
