package errorjson_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/errorjson"
)

// Example shows how errorjson.New serves as a terminal JSON error handler for an
// express application. It registers a route that forwards an error by calling
// next(err), then mounts errorjson.New last so that the propagated error is
// caught at the end of the chain. The request is driven with net/http/httptest,
// and both the resulting status code and the JSON body envelope are printed.
// Because the handler emits a compact, deterministic {"error": "..."} document
// with a fixed 500 status, the exact output is asserted in the Output comment.
func Example() {
	app := express.New()
	app.Get("/boom", func(req *express.Request, res *express.Response, next express.Next) {
		next(errors.New("something failed"))
	})
	app.Use(errorjson.New())

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 500
	// {"error":"something failed"}
}
