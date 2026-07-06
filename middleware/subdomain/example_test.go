package subdomain_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/subdomain"
)

// ExampleNew wires the subdomain middleware into an express application and
// drives it with net/http/httptest to show multi-tenant host routing. The
// middleware is configured with an explicit BaseHost of "example.com" and mounted
// at the front of the chain, so it computes the subdomain once and stores it on
// the request under subdomain.Key. A request addressed to api.example.com then has
// its ".example.com" suffix stripped, leaving "api" as the tenant label. The
// handler reads that value back with req.Value and captures it, and the example
// prints the captured subdomain after ServeHTTP completes. Setting BaseHost
// explicitly is recommended for real deployments because the fallback heuristic
// cannot know multi-label public suffixes.
func ExampleNew() {
	app := express.New()
	app.Use(subdomain.New(subdomain.Options{BaseHost: "example.com"}))

	var tenant string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		if v, ok := req.Value(subdomain.Key); ok {
			tenant, _ = v.(string)
		}
		res.Send("ok")
	})

	app.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "http://api.example.com/", nil))

	fmt.Println(tenant)
	// Output: api
}
