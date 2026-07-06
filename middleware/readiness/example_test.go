package readiness_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/readiness"
)

// ExampleNew wires the readiness probe into an application and drives it with
// httptest to show how the endpoint reflects application state. The middleware
// is registered globally via app.Use so it runs ahead of the router, and a
// closure-captured flag stands in for whatever real signal (cache warmed,
// migrations done) would gate traffic. The first probe is issued while the
// application reports itself not ready and therefore returns 503; the flag is
// then flipped and a second probe returns 200. Requests to other paths, such as
// "/", bypass the probe entirely and reach the normal route. The status codes
// and bodies are fully deterministic, so the example verifies them with an
// Output block.
func ExampleNew() {
	ready := false

	app := express.New()
	app.Use(readiness.New(readiness.Options{
		Path:  "/readyz",
		Ready: func() bool { return ready },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("home")
	})

	probe := func() (int, string) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/readyz", nil)
		app.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}

	code, body := probe()
	fmt.Printf("before: %d %s\n", code, body)

	ready = true
	code, body = probe()
	fmt.Printf("after: %d %s\n", code, body)

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Printf("route: %d %s\n", rec.Code, rec.Body.String())

	// Output:
	// before: 503 not ready
	// after: 200 ready
	// route: 200 home
}
