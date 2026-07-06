package requestdump_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/requestdump"
)

// ExampleNew wires the request-dump middleware into an express application and
// drives it with net/http/httptest to show how captured snapshots are read back.
// Because the capture ring is process-global and shared across the package, the
// example first calls Reset to isolate itself from any earlier traffic. The
// middleware is registered ahead of the route so it records each request before
// the handler runs, storing the method, path, headers, and capture time in a
// bounded ring. After ServeHTTP drives a GET to /users, Last returns the most
// recent snapshot, and the example prints its method and path to demonstrate the
// structured inspection this middleware is designed for.
func ExampleNew() {
	requestdump.Reset()

	app := express.New()
	app.Use(requestdump.New())
	app.Get("/users", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	app.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/users", nil))

	d := requestdump.Last()
	fmt.Println(d.Method, d.Path)
	// Output: GET /users
}
