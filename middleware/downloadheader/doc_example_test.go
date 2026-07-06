package downloadheader_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/downloadheader"
)

// Example demonstrates mounting the downloadheader middleware so that every
// response is served as a file download. It builds an express application,
// registers downloadheader.New with an Options value carrying a suggested
// filename, and adds a route that returns some CSV-like text. A request is then
// driven through the app with net/http/httptest, and the resulting
// Content-Disposition header is printed to show the attachment disposition and
// the quoted filename produced by the middleware. The deterministic header
// value is asserted through the Output comment below.
func Example() {
	app := express.New()
	app.Use(downloadheader.New(downloadheader.Options{Filename: "report.csv"}))
	app.Get("/export", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("id,name\n1,alice\n")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/export", nil))

	fmt.Println(rec.Header().Get("Content-Disposition"))
	// Output: attachment; filename="report.csv"
}
