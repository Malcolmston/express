package referer_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/referer"
)

// ExampleNew captures the Referer header once and reads it back downstream. The
// middleware is registered via app.Use so that the parsed value is available to
// every later handler through referer.From. A route then pulls the stored
// Referer and prints both its raw URL and its parsed Host, demonstrating that
// handlers consume a typed value instead of re-parsing the header. The request
// is driven through httptest with a Referer header set to a full URL including a
// path and query string. Because parsing is deterministic the example verifies
// the extracted URL and Host with an Output block.
func ExampleNew() {
	app := express.New()
	app.Use(referer.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		if ref, ok := referer.From(req); ok {
			fmt.Printf("url:  %s\n", ref.URL)
			fmt.Printf("host: %s\n", ref.Host)
		}
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Referer", "https://example.com/page?x=1")
	app.ServeHTTP(rec, req)

	// Output:
	// url:  https://example.com/page?x=1
	// host: example.com
}
