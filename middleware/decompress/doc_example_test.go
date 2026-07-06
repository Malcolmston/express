package decompress_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/decompress"
)

// Example sends a gzip-encoded request body through the decompress middleware and
// shows the handler reading it back as plain text. The client compresses a
// string with compress/gzip and sets Content-Encoding: gzip on the request, just
// as a bandwidth-conscious uploader would. The middleware detects the encoding,
// wraps the body in a gzip reader, strips the Content-Encoding header, and clears
// the content length, so the downstream handler's io.ReadAll returns the
// original, inflated bytes. The output is deterministic because the decompressed
// payload equals exactly what was compressed. This makes the middleware the
// inverse of response compression for incoming request bodies.
func Example() {
	app := express.New()
	app.Use(decompress.New())
	app.Post("/upload", func(req *express.Request, res *express.Response, next express.Next) {
		data, _ := io.ReadAll(req.Raw.Body)
		res.Send(string(data))
	})

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte("hello compressed world"))
	gw.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
	// Output: hello compressed world
}
