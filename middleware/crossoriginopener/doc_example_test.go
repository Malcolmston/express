package crossoriginopener_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/crossoriginopener"
)

// Example mounts the Cross-Origin-Opener-Policy middleware on an express
// application and drives it with net/http/httptest. It passes an explicit
// Policy of same-origin-allow-popups, which isolates the document from openers
// while still allowing it to reference popups it opens (useful for OAuth or
// payment flows). The route handler writes a body, demonstrating that the
// middleware sets its header and calls next() rather than short-circuiting.
// After serving the request we print the Cross-Origin-Opener-Policy response
// header. Because the header setter is deterministic, the expected output is
// asserted below.
func Example() {
	app := express.New()
	app.Use(crossoriginopener.New(crossoriginopener.Options{Policy: "same-origin-allow-popups"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Cross-Origin-Opener-Policy"))
	// Output:
	// same-origin-allow-popups
}
