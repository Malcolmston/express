package requestcontext_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/requestcontext"
)

// ExampleNew demonstrates attaching a request-scoped context to each request
// and retrieving it downstream with From. A deterministic Generator is supplied
// via Options so the example produces stable output; in production the default
// crypto/rand generator is used instead. The middleware is registered with
// app.Use before the route handler, so by the time the handler runs the *Ctx is
// already stored on the request and mirrored onto the X-Request-Id response
// header. The handler calls From(req) to read back the same ID, and the example
// drives the request through httptest and prints both the context ID and the
// response header to show they match.
func ExampleNew() {
	app := express.New()
	app.Use(requestcontext.New(requestcontext.Options{
		Generator: func() string { return "req-123" },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		ctx := requestcontext.From(req)
		res.Send(ctx.ID)
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println("ctx id:", rec.Body.String())
	fmt.Println("header:", rec.Header().Get(requestcontext.HeaderName))
	// Output:
	// ctx id: req-123
	// header: req-123
}
