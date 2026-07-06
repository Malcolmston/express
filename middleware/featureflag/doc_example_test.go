package featureflag_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/featureflag"
)

// Example wires the feature-flag middleware into an express application and
// drives it with net/http/httptest. It registers two flags at startup — one
// enabled and one disabled — via featureflag.Options, then mounts the
// middleware globally with app.Use so every request carries the flag set. The
// route handler queries the flags through featureflag.Enabled, which recovers
// the map the middleware stashed on the request. A flag that was never
// registered ("nope") reports false, demonstrating the graceful-degradation
// semantics. The printed lines are deterministic, so the Output block asserts
// the exact flag states seen by the handler.
func Example() {
	app := express.New()
	app.Use(featureflag.New(featureflag.Options{
		Flags: map[string]bool{"new-ui": true, "beta": false},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		fmt.Println("new-ui:", featureflag.Enabled(req, "new-ui"))
		fmt.Println("beta:", featureflag.Enabled(req, "beta"))
		fmt.Println("nope:", featureflag.Enabled(req, "nope"))
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	// Output:
	// new-ui: true
	// beta: false
	// nope: false
}
