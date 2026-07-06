package expectct_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/expectct"
)

// Example demonstrates configuring the expectct middleware with a full set of
// directives. It builds an express app and mounts expectct.New with an Options
// value that sets a max-age, enables enforcement, and supplies a report URI.
// A route is registered and a single request is driven through the app with
// net/http/httptest. The resulting Expect-CT header is printed to show the
// deterministic directive string the middleware assembles, and the exact value
// is asserted in the Output comment below.
func Example() {
	app := express.New()
	app.Use(expectct.New(expectct.Options{
		MaxAge:    86400,
		Enforce:   true,
		ReportURI: "https://example.com/report",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Expect-CT"))
	// Output: max-age=86400, enforce, report-uri="https://example.com/report"
}
