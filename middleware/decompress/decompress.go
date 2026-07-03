// Package decompress provides express middleware that transparently
// decompresses gzip-encoded request bodies so downstream handlers read plain
// data.
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

func (b *gzipBody) Read(p []byte) (int, error) { return b.gz.Read(p) }

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
