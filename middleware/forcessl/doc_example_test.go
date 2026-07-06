package forcessl_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/forcessl"
)

// Example mounts the force-SSL middleware at the top of an express application
// and drives it with net/http/httptest. It issues a plain HTTP request whose
// URL includes both a path and a query string, and shows that the middleware
// short-circuits the chain with a 301 redirect to the https equivalent rather
// than reaching the downstream handler. The Location header echoes the original
// host, path, and query string unchanged, so deep links survive the upgrade.
// Passing forcessl.New() with no options enables the redirect by default. The
// status code and Location value are fully deterministic, so the Output block
// asserts them exactly.
func Example() {
	app := express.New()
	app.Use(forcessl.New())
	app.Get("/path", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("should not reach here over http")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/path?x=1", nil)
	app.ServeHTTP(rec, req)

	fmt.Println("status:", rec.Code)
	fmt.Println("location:", rec.Header().Get("Location"))
	// Output:
	// status: 301
	// location: https://example.com/path?x=1
}
