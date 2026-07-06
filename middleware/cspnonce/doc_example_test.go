package cspnonce_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/cspnonce"
)

// Example wires the CSP nonce middleware into an express application and shows a
// route handler retrieving the per-request nonce with cspnonce.Nonce. On each
// request the middleware mints a fresh random nonce, stores it on the request
// and in res.Locals, and writes a Content-Security-Policy header that
// allow-lists that nonce in its script-src directive. A real handler would stamp
// the same nonce onto its inline <script> tags so the browser executes them. The
// nonce is random, so the example asserts structural facts rather than an exact
// value and omits an Output line because the token differs on every run.
func Example() {
	app := express.New()
	app.Use(cspnonce.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		nonce := cspnonce.Nonce(req)
		res.Send("<script nonce=\"" + nonce + "\">/* inline */</script>")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	csp := rec.Header().Get("Content-Security-Policy")
	fmt.Println(strings.HasPrefix(csp, "default-src 'self'; script-src 'self' 'nonce-"))
	// Output is intentionally omitted: the nonce is random on every request.
}
