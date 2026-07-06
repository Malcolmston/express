package cachecontrol_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/cachecontrol"
)

// ExampleNew mounts the cache-control middleware with a policy suitable for a
// long-lived static asset: "public" so shared caches may store it, together
// with a one-hour max-age. The example drives a single request through the
// application with httptest and prints the resulting Cache-Control response
// header. Directives are assembled in a stable order and joined with ", ", so
// the output is deterministic. The handler still runs and produces its body
// because the middleware only sets the header and then calls next() without
// short-circuiting the chain.
func ExampleNew() {
	app := express.New()
	app.Use(cachecontrol.New(cachecontrol.Options{
		Public: true,
		MaxAge: 3600,
	}))
	app.Get("/asset.js", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("// asset")
	})

	r := httptest.NewRequest("GET", "/asset.js", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Println(w.Header().Get("Cache-Control"))
	// Output:
	// public, max-age=3600
}
