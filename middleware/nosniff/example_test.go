package nosniff_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/nosniff"
)

// ExampleNew wires the nosniff middleware into an express application and shows
// that the X-Content-Type-Options header lands on the response. The middleware
// is registered globally with app.Use so every route inherits the header. A
// single route is added that returns a short body. The request is driven
// entirely in-memory with httptest.NewRequest and httptest.NewRecorder, then
// dispatched through app.ServeHTTP. Because the header value is a fixed constant
// the output is fully deterministic and can be asserted directly.
func ExampleNew() {
	app := express.New()
	app.Use(nosniff.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("X-Content-Type-Options"))
	// Output: nosniff
}
