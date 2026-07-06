package maintenance_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/maintenance"
)

// ExampleNew demonstrates gating an express application behind the maintenance
// middleware and flipping it at runtime through the returned Toggle. The
// handler is registered first with app.Use so it fronts every route, and a
// custom RetryAfter is configured so the emitted Retry-After header is
// deterministic. The example issues one request while maintenance mode is off,
// observing the normal 200 response from the route handler, then calls
// Toggle.Set(true) and issues a second request, observing the short-circuited
// 503 Service Unavailable together with its Retry-After header. Driving both
// requests with net/http/httptest recorders makes the before/after behavior
// fully reproducible.
func ExampleNew() {
	handler, toggle := maintenance.New(maintenance.Options{RetryAfter: 120})

	app := express.New()
	app.Use(handler)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	// Maintenance mode is off: the route responds normally.
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Printf("off:  status=%d\n", rec.Code)

	// Flip maintenance mode on: every request is short-circuited with 503.
	toggle.Set(true)
	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Printf("on:   status=%d retry-after=%s\n", rec.Code, rec.Header().Get("Retry-After"))

	// Output:
	// off:  status=200
	// on:   status=503 retry-after=120
}
