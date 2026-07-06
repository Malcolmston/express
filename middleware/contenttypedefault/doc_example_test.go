package contenttypedefault_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/contenttypedefault"
)

// Example demonstrates supplying a fallback Content-Type with the
// contenttypedefault middleware. It builds an express application and mounts
// the middleware configured with an explicit default media type. The handler
// then ends the response without declaring any Content-Type of its own, which
// is exactly the situation the middleware exists to cover. Just before the
// response headers are committed, the middleware's OnBeforeWrite callback
// notices the absent Content-Type and fills in the configured default. The
// example prints the header that the client would ultimately see, which is
// deterministic. Had the handler set its own type, the middleware would have
// left it untouched, because it only fills a genuinely missing value.
func Example() {
	app := express.New()
	app.Use(contenttypedefault.New(contenttypedefault.Options{
		Type: "application/octet-stream",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.End()
	})

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	fmt.Println(rr.Header().Get("Content-Type"))
	// Output: application/octet-stream
}
