package rewrite_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/rewrite"
)

// ExampleNew mounts the rewrite middleware ahead of the routes and drives a
// request whose public path no longer matches any handler directly. The single
// rule maps the legacy "/old/<rest>" space onto the current "/new/<rest>" space
// using a $1 capture-group substitution, and because SetPath updates the
// router's match path the request genuinely re-routes to the /new handler
// rather than 404ing. The example issues a GET for "/old/widgets", lets the
// middleware rewrite it internally, and prints the status and body produced by
// the handler registered under "/new/widgets". No client-visible redirect is
// involved, so the caller's URL is unchanged while the server routes the new
// path.
func ExampleNew() {
	app := express.New()

	app.Use(rewrite.New(rewrite.Options{
		Rules: []rewrite.Rule{
			{Pattern: `^/old/(.*)$`, To: "/new/$1"},
		},
	}))

	app.Get("/new/widgets", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("served by /new/widgets")
	})

	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest("GET", "/old/widgets", nil))
	fmt.Printf("%d %s\n", w.Code, w.Body.String())

	// Output:
	// 200 served by /new/widgets
}
