package referrerpolicy_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/referrerpolicy"
)

// ExampleNew demonstrates mounting the referrerpolicy middleware on an
// application so that every response carries a Referrer-Policy header. Here a
// custom policy list is supplied via Options, and the tokens are joined with
// ", " to form the header value that browsers use as an ordered fallback. The
// middleware is registered with app.Use before the route handler so the header
// is set on the outgoing response regardless of what the handler does. The
// request is driven through httptest so no real network is involved, and the
// resulting header is printed. Because the value is fixed for a given
// configuration, the output is deterministic.
func ExampleNew() {
	app := express.New()
	app.Use(referrerpolicy.New(referrerpolicy.Options{
		Policy: []string{"no-referrer", "strict-origin-when-cross-origin"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Referrer-Policy"))
	// Output: no-referrer, strict-origin-when-cross-origin
}
