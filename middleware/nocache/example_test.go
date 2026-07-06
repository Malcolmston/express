package nocache_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/nocache"
)

// ExampleNew registers the nocache middleware on an express application so that
// its anti-caching headers are applied to every response. The middleware is
// installed globally with app.Use, ahead of the route handlers. A single route
// returns a dynamic body that should never be cached by clients or proxies. The
// request is exercised in-memory with httptest.NewRequest and a recorder, then
// dispatched via app.ServeHTTP. Because the three header values are fixed
// constants the output is deterministic and asserted directly.
func ExampleNew() {
	app := express.New()
	app.Use(nocache.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("dynamic")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Cache-Control"))
	fmt.Println(rec.Header().Get("Pragma"))
	fmt.Println(rec.Header().Get("Expires"))
	// Output:
	// no-store, no-cache, must-revalidate
	// no-cache
	// 0
}
