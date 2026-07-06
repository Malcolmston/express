package xssfilter_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/xssfilter"
)

// ExampleNew wires the xss-filter middleware into an express application and
// drives it with net/http/httptest. The middleware is mounted at the front of the
// chain so the X-XSS-Protection header is written on every response before the
// route handler runs. With the zero-value options used here the header takes its
// default value "0", which disables the legacy reflected-XSS auditor in browsers
// that still honor it — the modern, recommended posture. After ServeHTTP the
// example reads the committed header from the recorder and prints it. Real XSS
// defense should come from a Content-Security-Policy and correct output encoding
// rather than from this header.
func ExampleNew() {
	app := express.New()
	app.Use(xssfilter.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	fmt.Println(rec.Header().Get("X-XSS-Protection"))
	// Output: 0
}
