// Package decompress provides express middleware that transparently
// decompresses gzip-encoded request bodies so downstream handlers read plain
// data. It is the inverse of response compression: where a compression
// middleware shrinks outgoing responses, decompress inflates incoming requests,
// mirroring the request-body inflation that Node's HTTP stack and middleware
// such as body-parser perform when a client sends a gzip-encoded payload.
//
// Use this middleware when clients — mobile apps, IoT devices, or services
// uploading large JSON, logs, or telemetry — send request bodies with
// "Content-Encoding: gzip" to save bandwidth. Without it, a handler reading
// req.Raw.Body would see raw compressed bytes; with it, the handler reads the
// original plain data as though no encoding had been applied. Mount it with
// app.Use ahead of any body-parsing middleware or route handler that consumes
// the request body, so the body is already inflated by the time they read it.
//
// Operationally the middleware inspects the Content-Encoding request header on
// each request. When the value (trimmed and lower-cased) is "gzip" or its alias
// "x-gzip", it wraps req.Raw.Body in a gzip.Reader whose Close also closes the
// original body, removes the Content-Encoding header, and sets ContentLength to
// -1 to signal that the inflated length is unknown. It then calls next() so the
// request proceeds with a body that streams plain data on demand. Requests with
// any other encoding, or with no body, are passed through untouched with a plain
// next().
//
// The failure mode is explicit: gzip.NewReader validates the gzip header
// immediately, so if the declared encoding is gzip but the bytes are not a valid
// gzip stream, the middleware calls next(err) to hand the error to the
// framework's error pipeline (which yields a 500 by default) rather than letting
// a corrupt stream reach the handler. Decompression is streaming, so a truncated
// body surfaces its error later, when the handler reads to the end. Because
// ContentLength is cleared, downstream code must not rely on a fixed body length
// after this middleware runs.
//
// This is a Go-native port rather than a wrapper of a specific npm package, and
// it is deliberately minimal. It handles only the gzip family — not deflate,
// br (Brotli), or other content codings — applies no size or ratio limit to
// guard against decompression-bomb amplification, and does not attempt to
// re-derive the true Content-Length. Callers who need those protections should
// enforce an upstream request-size limit before this middleware and add codec
// support as needed.
package decompress

import (
	"compress/gzip"
	"io"
	"strings"

	"github.com/malcolmston/express"
)

// gzipBody couples the gzip.Reader with the underlying body so both are closed.
type gzipBody struct {
	gz   *gzip.Reader
	orig io.ReadCloser
}

// Read implements io.Reader; it reads decompressed data from the gzip reader.
func (b *gzipBody) Read(p []byte) (int, error) { return b.gz.Read(p) }

// Close implements io.Closer; it closes the gzip reader and the underlying
// body, returning the first error encountered.
func (b *gzipBody) Close() error {
	err := b.gz.Close()
	if cerr := b.orig.Close(); err == nil {
		err = cerr
	}
	return err
}

// New returns middleware that, when the request's Content-Encoding is gzip,
// wraps req.Raw.Body in a gzip reader. The Content-Encoding header is removed
// and Content-Length cleared so downstream code treats the body as plain,
// unknown-length data. Non-gzip requests pass through untouched.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		enc := strings.ToLower(strings.TrimSpace(req.Get("Content-Encoding")))
		if enc != "gzip" && enc != "x-gzip" {
			next()
			return
		}
		if req.Raw.Body == nil {
			next()
			return
		}
		gz, err := gzip.NewReader(req.Raw.Body)
		if err != nil {
			next(err)
			return
		}
		req.Raw.Body = &gzipBody{gz: gz, orig: req.Raw.Body}
		req.Raw.Header.Del("Content-Encoding")
		req.Raw.ContentLength = -1
		next()
	}
}
