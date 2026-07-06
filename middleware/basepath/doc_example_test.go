package basepath_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/basepath"
)

// ExampleNew demonstrates serving routes authored at root paths while the
// application is mounted under a "/app" prefix. The middleware is mounted first
// with app.Use so the prefix is stripped before routing, letting the "/users"
// route be registered without any prefix. A request to "/app/users" therefore
// matches the "/users" handler. The rewritten path is deterministic, so an
// Output block asserts the handler ran and observed the stripped path.
func ExampleNew() {
	app := express.New()
	app.Use(basepath.New(basepath.Options{Prefix: "/app"}))
	app.Get("/users", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("matched " + req.Path())
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/app/users", nil))

	fmt.Println(rec.Body.String())
	// Output: matched /users
}
