package redirectmap_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/redirectmap"
)

// ExampleNew builds a static redirect table and drives two requests through it.
// The middleware is constructed with a Map of old-to-new paths and a custom 301
// status so that matched requests are permanently redirected. It is registered
// via app.Use ahead of a catch-all route that stands in for the rest of the
// application. The first request targets a mapped path ("/old") and is
// short-circuited into a redirect carrying the Location header; the second
// targets an unmapped path ("/keep") and falls through to the route. The status
// codes and Location value are deterministic, so the example asserts its
// Output.
func ExampleNew() {
	app := express.New()
	app.Use(redirectmap.New(redirectmap.Options{
		Map:    map[string]string{"/old": "/new"},
		Status: http.StatusMovedPermanently,
	}))
	app.Get("/keep", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("kept")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/old", nil))
	fmt.Printf("redirect: %d -> %s\n", rec.Code, rec.Header().Get("Location"))

	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/keep", nil))
	fmt.Printf("fallthrough: %d %s\n", rec.Code, rec.Body.String())

	// Output:
	// redirect: 301 -> /new
	// fallthrough: 200 kept
}
