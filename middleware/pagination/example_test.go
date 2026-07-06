package pagination_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/pagination"
)

// ExampleNew shows how a list endpoint consumes parsed paging parameters. The
// middleware is configured with a default and maximum limit and registered
// globally via app.Use, so it parses ?page and ?limit once per request before
// the route runs. Inside the handler, pagination.From retrieves the normalized
// Pagination value, whose Page, Limit, and derived Offset are ready to feed a
// database LIMIT/OFFSET query. The example issues a request with an
// out-of-range limit to demonstrate clamping to MaxLimit, then drives it through
// app.ServeHTTP with an httptest recorder and prints the resulting values so the
// deterministic output can be asserted.
func ExampleNew() {
	app := express.New()
	app.Use(pagination.New(pagination.Options{DefaultLimit: 20, MaxLimit: 50}))
	app.Get("/items", func(req *express.Request, res *express.Response, next express.Next) {
		p := pagination.From(req)
		res.JSON(map[string]int{
			"page":   p.Page,
			"limit":  p.Limit,
			"offset": p.Offset,
		})
	})

	// limit=999 is clamped to MaxLimit (50); offset is (page-1)*limit.
	req := httptest.NewRequest(http.MethodGet, "/items?page=3&limit=999", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
	// Output: {"limit":50,"offset":100,"page":3}
}
