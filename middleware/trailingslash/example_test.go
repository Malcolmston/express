package trailingslash_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/trailingslash"
)

// ExampleNew demonstrates enforcing a trailing-slash canonicalization policy.
// It mounts the middleware with Enforce set on an express.Application ahead of
// a catch-all handler, then drives three requests through httptest to exercise
// the distinct paths: a slashless "/about" is redirected to the canonical
// "/about/", an already-conforming "/about/" passes straight through to the
// handler, and a slashless path with a query string keeps that query on the
// redirect target. The root path is always left untouched by design. Because
// the default status is 301 and the outcomes depend only on the requested URL,
// the Output block is deterministic.
func ExampleNew() {
	app := express.New()
	app.Use(trailingslash.New(trailingslash.Options{Enforce: true}))
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	call := func(path string) {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		if loc := w.Header().Get("Location"); loc != "" {
			fmt.Printf("%-12s -> %d %s\n", path, w.Code, loc)
		} else {
			fmt.Printf("%-12s -> %d %s\n", path, w.Code, w.Body.String())
		}
	}
	call("/about")
	call("/about/")
	call("/report?y=1")
	// Output:
	// /about       -> 301 /about/
	// /about/      -> 200 ok
	// /report?y=1  -> 301 /report/?y=1
}
