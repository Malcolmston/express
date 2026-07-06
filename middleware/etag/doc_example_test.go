package etag_test

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/etag"
)

// Example demonstrates the conditional-request behavior of the etag middleware.
// It builds an express app, mounts etag.New ahead of a route that sends a fixed
// body, and issues two requests through net/http/httptest. The first request
// carries no validator, so the middleware returns 200 with the body and an ETag
// header; the example recomputes that expected tag with crypto/sha1 to prove it
// matches. The second request replays the tag in If-None-Match, so the
// middleware short-circuits with a 304 Not Modified and an empty body, and the
// deterministic status codes are asserted via the Output comment.
func Example() {
	app := express.New()
	app.Use(etag.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("hello")
	})

	// First request: no validator, full body plus ETag.
	rec1 := httptest.NewRecorder()
	app.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/", nil))
	tag := rec1.Header().Get("ETag")

	sum := sha1.Sum([]byte("hello"))
	expected := `"` + hex.EncodeToString(sum[:]) + `"`
	fmt.Println(rec1.Code, tag == expected)

	// Second request: replay the tag, expect a 304 with no body.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("If-None-Match", tag)
	rec2 := httptest.NewRecorder()
	app.ServeHTTP(rec2, req2)
	fmt.Println(rec2.Code, rec2.Body.Len())

	// Output:
	// 200 true
	// 304 0
}
