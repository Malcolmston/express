package referercheck_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/referercheck"
)

// ExampleNew guards a route with a Referer allowlist and drives three requests
// through it. The middleware is constructed with a single permitted host and
// registered via app.Use ahead of the protected route, so it decides whether
// each request reaches the handler. The first request carries a Referer from the
// allowed host and is admitted with 200; the second carries a Referer from a
// disallowed host and is rejected with 403; the third omits the Referer entirely
// and is also rejected, because Optional is left at its default of false. The
// status codes are deterministic, so the example asserts them with an Output
// block.
func ExampleNew() {
	app := express.New()
	app.Use(referercheck.New(referercheck.Options{
		Allow: []string{"example.com"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	probe := func(ref string) int {
		r := httptest.NewRequest("GET", "/", nil)
		if ref != "" {
			r.Header.Set("Referer", ref)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	fmt.Println("allowed:   ", probe("https://example.com/page"))
	fmt.Println("disallowed:", probe("https://evil.com/page"))
	fmt.Println("missing:   ", probe(""))

	// Output:
	// allowed:    200
	// disallowed: 403
	// missing:    403
}
