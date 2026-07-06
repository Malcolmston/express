package compression_test

import (
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/compression"
)

// Example demonstrates gzip-compressing responses with the compression
// middleware. It builds an express application, mounts the middleware with a
// low MinLength so even a modest body qualifies, and registers a handler that
// sends a repetitive, highly compressible payload. The request advertises gzip
// support through its Accept-Encoding header, which is the signal the
// middleware negotiates on. Because the body exceeds MinLength and carries no
// existing Content-Encoding, the middleware buffers it, gzips it, and sets
// Content-Encoding: gzip together with Vary: Accept-Encoding on the response.
// The example then prints the negotiated Content-Encoding header, which is
// deterministic regardless of the exact compressed bytes.
func Example() {
	app := express.New()
	app.Use(compression.New(compression.Options{MinLength: 16}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(strings.Repeat("hello world ", 64))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	fmt.Println(rr.Header().Get("Content-Encoding"))
	// Output: gzip
}
