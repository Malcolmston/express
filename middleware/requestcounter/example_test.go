package requestcounter_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/requestcounter"
)

// ExampleNew demonstrates the two-value shape of the requestcounter
// constructor: it returns both the middleware handler and an accessor that
// reports how many requests have been observed so far. The handler is mounted
// with app.Use so that every request increments the shared, atomic counter
// before reaching the route. Several requests are then driven through the app
// with httptest, and the accessor is queried afterward to read the running
// total. The accessor performs an atomic load and is safe to call from other
// goroutines. No Output directive is used here to keep the example robust, but
// after five requests the accessor reports five.
func ExampleNew() {
	handler, count := requestcounter.New()

	app := express.New()
	app.Use(handler)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	fmt.Printf("handled %d requests\n", count())
}
