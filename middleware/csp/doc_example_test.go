package csp_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/csp"
)

// Example builds a Content-Security-Policy middleware from a directive map and
// mounts it on an express application. The middleware sets the header on every
// response before the route handler runs, so the policy is present regardless of
// what the handler writes. Directive names are emitted in sorted order, giving a
// deterministic header value that is safe to assert against. Here we allow
// scripts from the same origin plus a CDN while keeping the default source
// restricted to 'self'. The example drives the app with httptest and prints the
// resulting header so the exact policy string is visible.
func Example() {
	app := express.New()
	app.Use(csp.New(csp.Options{Directives: map[string][]string{
		"default-src": {"'self'"},
		"script-src":  {"'self'", "https://cdn.example.com"},
	}}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Content-Security-Policy"))
	// Output: default-src 'self'; script-src 'self' https://cdn.example.com
}
