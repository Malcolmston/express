package notfound_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/notfound"
)

// ExampleNew shows the notfound handler acting as the terminal 404 responder for
// unmatched routes. A real route is registered first, then notfound.New is added
// last with app.Use so it only runs when nothing earlier handled the request.
// The example issues a request to a path that no route matches, driving it
// in-memory with httptest.NewRequest and a recorder through app.ServeHTTP. The
// handler sets a 404 status and writes the default "Not Found" body without
// calling next. Both the status code and body are deterministic and printed for
// assertion.
func ExampleNew() {
	app := express.New()
	app.Get("/home", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("home")
	})
	app.Use(notfound.New())

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 404
	// Not Found
}
