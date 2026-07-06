package responsetime_test

import (
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/responsetime"
)

// ExampleNew mounts the response-time middleware at the front of an express app
// and drives a request through it with httptest. The middleware records a start
// time, registers an OnBeforeWrite hook, and lets the handler run; when the
// response headers are committed the hook stamps the elapsed duration onto the
// X-Response-Time header (whose name is exported as HeaderName). Because the
// measured value depends on wall-clock timing it is not deterministic, so this
// example asserts on the header's shape rather than its exact contents. It
// prints whether the header was set and whether its value carries the expected
// "ms" suffix, both of which are stable regardless of how fast the handler ran.
// For that reason the example intentionally omits an Output comment.
func ExampleNew() {
	app := express.New()
	app.Use(responsetime.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	got := w.Header().Get(responsetime.HeaderName)
	fmt.Printf("header set: %t\n", got != "")
	fmt.Printf("has ms suffix: %t\n", strings.HasSuffix(got, "ms"))
}
