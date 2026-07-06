package vhost_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/vhost"
)

// ExampleNew wires the virtual-host middleware into an express application and
// drives it with net/http/httptest to show host-based dispatch. The middleware is
// configured to match the hostname "api.example.com" and, on a match, to invoke a
// dedicated handler that serves the API site's response. It is mounted near the
// top of the chain so host dispatch happens before the generic route registered
// with app.Get. A request whose Host header is api.example.com therefore never
// reaches the main-site handler; instead the vhost handler responds with its own
// body, which the example reads back from the recorder. A request for any other
// host would fall through to next() and be served by the main route.
func ExampleNew() {
	app := express.New()
	app.Use(vhost.New(vhost.Options{
		Host: "api.example.com",
		Handler: func(req *express.Request, res *express.Response, next express.Next) {
			res.Send("api response")
		},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("main site")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "http://api.example.com/", nil))

	fmt.Println(rec.Body.String())
	// Output: api response
}
