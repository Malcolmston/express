package sanitize_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/sanitize"
)

// Example demonstrates the sanitize middleware cleaning a query parameter that
// carries an injected <script> tag. The middleware is mounted first so it runs
// ahead of the route handler, stripping HTML tags from every query value in
// place and rewriting the request URL. By the time the handler reads the "name"
// parameter with req.Query, the markup is gone and only the inner text remains,
// so a value that is later echoed into a page cannot smuggle a live element.
// The request is never short-circuited: sanitize always calls next and lets the
// handler produce its normal response.
func Example() {
	app := express.New()
	app.Use(sanitize.New())

	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("Hello, " + req.Query("name"))
	})

	// name=<script>alert(1)</script>Ann, URL-encoded.
	r := httptest.NewRequest("GET", "/?name=%3Cscript%3Ealert(1)%3C%2Fscript%3EAnn", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Println(w.Body.String())
	// Output:
	// Hello, alert(1)Ann
}
