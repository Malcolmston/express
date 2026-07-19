// Package etag provides express middleware that computes a strong ETag for the
// response body and answers conditional requests with 304 Not Modified. It
// ports the response-side ETag behavior that Express enables by default via the
// Node "etag" package together with fresh/If-None-Match handling: a validator
// derived from the body is emitted, and a matching conditional request is
// answered with an empty 304 rather than the full payload.
//
// Use this middleware to cut bandwidth and speed up repeat requests for
// responses whose bytes are stable between requests — rendered pages, JSON API
// results, static-ish content generated per request. It is most valuable in
// front of handlers that regenerate identical output, letting caches and
// browsers revalidate cheaply with If-None-Match instead of re-downloading the
// body. Mount it with app.Use so it wraps the routes whose responses you want
// to tag.
//
// Mechanically the middleware must sit upstream of the handlers whose output it
// tags, because it works by buffering. On each request it swaps res.Writer for
// an internal captureWriter, calls next() to let the downstream chain run,
// then restores the original writer. The captureWriter records the status code
// and accumulates all Write calls into an in-memory buffer instead of sending
// them, so nothing reaches the client until the middleware decides what to do.
// This means the entire response body is held in memory; it is unsuitable for
// very large or streaming responses.
//
// After next() returns, the buffered body drives the outcome. If the body is
// empty the middleware writes through only the status with no ETag (there is
// nothing to validate). Otherwise it computes the tag as the SHA-1 of the body,
// hex-encoded and wrapped in double quotes, and sets it as the ETag header —
// this is a strong validator, with no W/ weak prefix. It then reads the
// request's If-None-Match via req.Get. On a match it deletes the Content-Type
// and Content-Length headers and sends 304 Not Modified with an empty body; on
// a miss it writes the captured status followed by the buffered bytes unchanged.
// If-None-Match matching treats "*" as matching any tag, otherwise splits the
// header on commas, trims each entry, strips a leading W/ weak-validator prefix,
// and compares for equality against the computed strong tag.
//
// Regarding parity with the Node original: the default Express/etag behavior
// hashes the payload (SHA-1 based) and emits a quoted, hex validator, and
// Express answers fresh conditional GETs with 304 — both of which this port
// reproduces. The notable differences are that the Node etag package emits a
// weak validator (W/-prefixed) by default and encodes the digest in base64
// including the byte length, whereas this port always emits a strong,
// hex-encoded SHA-1 tag; and that it only consults If-None-Match, not
// If-Modified-Since or Last-Modified. It also does not skip tagging based on
// status code, so error and redirect bodies are tagged too as long as they are
// non-empty.
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

// WriteHeader implements http.ResponseWriter; it records the first status code
// without sending it so the header can be written later once the full buffered
// payload is available.
func (w *captureWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
}

// Write implements http.ResponseWriter; it buffers p in memory (defaulting the
// status to 200 on first write) rather than writing to the client immediately.
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
