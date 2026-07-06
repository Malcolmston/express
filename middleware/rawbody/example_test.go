package rawbody_test

import (
	"fmt"
	"io"
	"net/http/httptest"
	"strings"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/rawbody"
)

// ExampleNew demonstrates buffering the request body so it can be read more
// than once. The middleware is registered with app.Use ahead of the route so
// it drains the incoming stream into memory before the handler runs. A POST
// route then retrieves the captured bytes with req.Body (typed as a []byte)
// and independently re-reads req.Raw.Body, which the middleware restored to a
// fresh reader over the same bytes. The request is driven through the app with
// httptest carrying a small text payload. The example prints both the captured
// buffer and the re-read stream, showing they are identical and that the body
// survived being consumed by the middleware.
func ExampleNew() {
	app := express.New()
	app.Use(rawbody.New())
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		captured, _ := req.Body().([]byte)
		reread, _ := io.ReadAll(req.Raw.Body)
		fmt.Printf("captured=%q reread=%q\n", captured, reread)
		res.Send("ok")
	})

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("POST", "/", strings.NewReader("hello world")))
	// Output: captured="hello world" reread="hello world"
}
