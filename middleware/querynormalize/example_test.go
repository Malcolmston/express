package querynormalize_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/querynormalize"
)

// ExampleNew demonstrates canonicalizing the request query string before it
// reaches route handlers. The middleware is registered with app.Use so it runs
// first, rewriting req.Raw.URL.RawQuery in place. A GET route then reads the
// already-normalized query and echoes it back. The request is driven through
// the app with httptest using a query whose keys are upper-cased, out of order,
// and padded with an encoded space. After ServeHTTP returns, the handler has
// observed lower-cased keys sorted alphabetically with values trimmed, which
// the example prints as deterministic output.
func ExampleNew() {
	app := express.New()
	app.Use(querynormalize.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(req.Raw.URL.RawQuery)
	})

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/?B=2&A=%20bob%20", nil))

	fmt.Println(rr.Body.String())
	// Output: a=bob&b=2
}
