package crossoriginresource_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/crossoriginresource"
)

// Example mounts the Cross-Origin-Resource-Policy middleware on an express
// application and drives it with net/http/httptest. It passes an explicit
// Policy of cross-origin, which is appropriate for a public asset that any
// origin should be allowed to embed (for example a CDN image or a shared API
// response). The route handler writes a body, demonstrating that the
// middleware sets its header and calls next() rather than short-circuiting.
// After serving the request we print the Cross-Origin-Resource-Policy response
// header. Because the header setter is deterministic, the expected output is
// asserted below.
func Example() {
	app := express.New()
	app.Use(crossoriginresource.New(crossoriginresource.Options{Policy: "cross-origin"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Cross-Origin-Resource-Policy"))
	// Output:
	// cross-origin
}
