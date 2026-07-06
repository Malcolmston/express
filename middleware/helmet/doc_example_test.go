package helmet_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/helmet"
)

// Example wires the helmet middleware into an express application and drives it
// with net/http/httptest. A single app.Use installs the whole bundle of default
// security headers, and this example overrides two of them via Options by
// selecting the DENY frameguard action and enabling includeSubDomains on HSTS.
// The middleware also composes hidepoweredby, so it strips X-Powered-By at write
// time even though the framework would otherwise emit it. Because helmet sets
// several headers at once, printing them all would be brittle, so the example
// asserts on exactly one deterministic header — X-Frame-Options — to keep the
// Output block stable. The other headers (nosniff, HSTS, Referrer-Policy, and
// the rest) are set on the same response and can be inspected the same way.
func Example() {
	app := express.New()
	app.Use(helmet.New(helmet.Options{
		FrameguardAction:      "DENY",
		HSTSIncludeSubDomains: true,
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println("X-Frame-Options:", rec.Header().Get("X-Frame-Options"))
	// Output:
	// X-Frame-Options: DENY
}
