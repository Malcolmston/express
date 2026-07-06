package flash_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/flash"
)

// Example demonstrates adding and consuming one-time flash messages within a
// single request, driven by net/http/httptest. It first mounts the express
// session middleware (required — the flash helpers operate on req.Session())
// and then the flash middleware itself. The handler records two messages under
// different categories with flash.Add, then drains them with flash.Get, which
// returns every pending message and clears the queue. A second flash.Get in the
// same request confirms the queue is now empty, illustrating the read-and-clear
// semantics that make flash messages appear exactly once. The output is
// deterministic because insertion order is preserved, so the Output block
// asserts the exact messages seen.
func Example() {
	app := express.New()
	app.Use(express.Session())
	app.Use(flash.New())

	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		flash.Add(req, "info", "profile saved")
		flash.Add(req, "error", "avatar too large")

		for _, m := range flash.Get(req) {
			fmt.Printf("%s: %s\n", m.Category, m.Message)
		}
		fmt.Println("second read empty:", flash.Get(req) == nil)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	// Output:
	// info: profile saved
	// error: avatar too large
	// second read empty: true
}
