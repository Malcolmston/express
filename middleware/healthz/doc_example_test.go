package healthz_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/healthz"
)

// Example mounts the healthz middleware on an express application and probes it
// with net/http/httptest. It calls healthz.New() with no options, so the
// endpoint defaults to serving the body "ok" at the path "/healthz". A GET to
// that path short-circuits the chain and returns a plain 200 response without
// reaching any downstream handler. A request to any other path falls through
// instead, which is what lets healthz coexist with normal routes. Both the
// status code and the body are fixed, so the Output block asserts them exactly.
func Example() {
	app := express.New()
	app.Use(healthz.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("home")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	fmt.Println("status:", rec.Code)
	fmt.Println("body:", rec.Body.String())
	// Output:
	// status: 200
	// body: ok
}
