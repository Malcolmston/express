package csrf_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/csrf"
)

// Example demonstrates the double-submit-cookie flow end to end. A first GET
// request causes the middleware to issue a token cookie, which the handler also
// exposes via csrf.Token so a page could embed it in a form. A second POST
// request then replays that token in both the cookie and the X-CSRF-Token
// header; because the two match, the middleware calls next and the protected
// handler runs. The token itself is random, so the example asserts only that the
// authorized POST succeeds and omits an Output line for the token value. This
// mirrors how a browser would resubmit the cookie automatically while a script
// copies the token into the request header.
func Example() {
	app := express.New()
	app.Use(csrf.New())
	app.Get("/form", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(csrf.Token(req))
	})
	app.Post("/submit", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("accepted")
	})

	// Simulate a client that already holds a token and submits it correctly.
	const token = "example-token-value"
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.AddCookie(&http.Cookie{Name: "csrf", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 200
	// accepted
}
