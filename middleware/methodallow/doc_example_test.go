package methodallow_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/methodallow"
)

// ExampleNew demonstrates guarding an express application so it only accepts a
// fixed set of HTTP methods. The middleware is registered first with app.Use
// and configured to permit GET and POST, while the route is mounted with
// app.All so that method rejection is enforced by the guard rather than by
// routing. The example drives two requests through net/http/httptest: an
// allowed POST that reaches the handler and returns 200, and a disallowed
// DELETE that is short-circuited with 405 and an Allow header advertising the
// permitted verbs. Printing the status codes and the Allow header shows both
// the pass-through and the rejection paths deterministically.
func ExampleNew() {
	app := express.New()
	app.Use(methodallow.New(methodallow.Options{Methods: []string{"GET", "POST"}}))
	app.All("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	// Permitted method: reaches the handler.
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("POST", "/", nil))
	fmt.Printf("POST:   status=%d\n", rec.Code)

	// Disallowed method: 405 with an Allow header.
	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("DELETE", "/", nil))
	fmt.Printf("DELETE: status=%d allow=%s\n", rec.Code, rec.Header().Get("Allow"))

	// Output:
	// POST:   status=200
	// DELETE: status=405 allow=GET, POST
}
