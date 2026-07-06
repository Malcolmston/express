package jsonp_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/jsonp"
)

// Example demonstrates how the jsonp middleware upgrades an ordinary JSON
// response into a JavaScript callback invocation when the request supplies a
// callback query parameter. It builds an Express application, mounts jsonp.New()
// globally with app.Use, and registers a route whose handler simply calls
// res.JSON with a map. The request is driven in-process with net/http/httptest
// using the query string "?callback=cb", so the middleware wraps the body as
// cb({...}); and rewrites the Content-Type to application/javascript. The
// example prints both the transformed body and the resulting Content-Type
// header, and because the JSON body is a single deterministic key the Output
// block is fully reproducible.
func Example() {
	app := express.New()
	app.Use(jsonp.New())
	app.Get("/data", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON(map[string]int{"a": 1})
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/data?callback=cb", nil))

	fmt.Println(rec.Body.String())
	fmt.Println(rec.Header().Get("Content-Type"))
	// Output:
	// cb({"a":1});
	// application/javascript; charset=utf-8
}
