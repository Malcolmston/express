package permittedcrossdomain_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/permittedcrossdomain"
)

// ExampleNew shows how to install the X-Permitted-Cross-Domain-Policies
// middleware with an explicit policy. Here Options.Policy is set to "master-only"
// so only the master crossdomain policy file at the domain root is honored by
// legacy Adobe clients; omitting Options, or leaving Policy empty, would default
// to the more restrictive "none". The middleware is registered globally via
// app.Use, so every response carries the header regardless of which route
// handles the request. The example records a GET request through app.ServeHTTP
// with an httptest recorder and prints the header value so the deterministic
// result can be asserted.
func ExampleNew() {
	app := express.New()
	app.Use(permittedcrossdomain.New(permittedcrossdomain.Options{
		Policy: "master-only",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("hello")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Header().Get("X-Permitted-Cross-Domain-Policies"))
	// Output: master-only
}
