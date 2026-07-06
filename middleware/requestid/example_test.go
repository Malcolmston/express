package requestid_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/requestid"
)

// ExampleNew registers the request-id middleware into an express application and
// drives it with net/http/httptest. The middleware is mounted first so that every
// request is assigned an identifier before any handler or logger runs; that id is
// echoed on the response under the configured header and stored on the request so
// downstream code can read it via req.Value(ContextKey). The handler here fetches
// the stored id purely to show it is available. Because the default generator
// draws a cryptographically random 16-byte value, the concrete id differs on
// every run, so this example asserts on the id's length rather than printing it
// and therefore omits an Output comment. In real use the same value ties together
// application logs and the client-visible response header.
func ExampleNew() {
	app := express.New()
	app.Use(requestid.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		if v, ok := req.Value(requestid.ContextKey); ok {
			_ = v // the id is available to loggers and handlers here
		}
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	id := rec.Header().Get(requestid.DefaultHeader)
	fmt.Println(len(id) == 32)
}

// ExampleNew_reuse shows that an inbound id is trusted and echoed verbatim rather
// than regenerated. An upstream proxy or gateway commonly stamps X-Request-Id so
// that a single identifier flows across every hop of a distributed system; this
// middleware honors that value when the configured header is already present on
// the incoming request. Here the request carries "trace-abc" in the default
// header, and the middleware mirrors exactly that onto the response instead of
// minting a fresh id. This behaviour is unconditional, so untrusted clients should
// have the header stripped or overridden upstream.
func ExampleNew_reuse() {
	app := express.New()
	app.Use(requestid.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set(requestid.DefaultHeader, "trace-abc")
	app.ServeHTTP(rec, r)

	fmt.Println(rec.Header().Get(requestid.DefaultHeader))
	// Output: trace-abc
}
