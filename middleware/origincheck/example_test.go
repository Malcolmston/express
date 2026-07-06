package origincheck_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/origincheck"
)

// ExampleNew demonstrates guarding a state-changing route with the origin-check
// middleware. The middleware is constructed with an allowlist of trusted origin
// hosts and registered globally via app.Use, so it runs before the route
// handler on every request. A request whose Origin header matches the allowlist
// is allowed through and reaches the handler, while a request from an untrusted
// origin is short-circuited with a 403 Forbidden response and never runs the
// handler. The example drives both cases through app.ServeHTTP with httptest
// recorders and prints each resulting status code to make the behavior
// observable and deterministic.
func ExampleNew() {
	app := express.New()
	app.Use(origincheck.New(origincheck.Options{
		Allow: []string{"example.com", "app.example.com:8443"},
	}))
	app.Post("/transfer", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	// Trusted origin: allowed through to the handler.
	good := httptest.NewRequest("POST", "/transfer", nil)
	good.Header.Set("Origin", "https://example.com")
	goodRec := httptest.NewRecorder()
	app.ServeHTTP(goodRec, good)

	// Untrusted origin: rejected with 403 before the handler runs.
	bad := httptest.NewRequest("POST", "/transfer", nil)
	bad.Header.Set("Origin", "https://evil.example.net")
	badRec := httptest.NewRecorder()
	app.ServeHTTP(badRec, bad)

	fmt.Println(goodRec.Code)
	fmt.Println(badRec.Code)
	// Output:
	// 200
	// 403
}
