package hsts_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/hsts"
)

// Example wires the hsts middleware into an express application and drives it
// with net/http/httptest. It configures a one-year max-age together with the
// includeSubDomains and preload directives so the emitted header value is fully
// deterministic. The middleware is registered with app.Use so it runs before
// the route handler and stamps the Strict-Transport-Security header onto every
// response. After serving a request we read the header back off the recorder to
// show exactly what a browser would receive. Because the header value is
// computed once at construction time, every response for this app carries the
// identical directive string.
func Example() {
	app := express.New()
	app.Use(hsts.New(hsts.Options{
		MaxAge:            31536000,
		IncludeSubDomains: true,
		Preload:           true,
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Strict-Transport-Security"))
	// Output:
	// max-age=31536000; includeSubDomains; preload
}
