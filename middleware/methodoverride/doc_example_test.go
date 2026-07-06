package methodoverride_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/methodoverride"
)

// ExampleNew demonstrates letting an HTML-form-style POST emulate a DELETE by
// carrying the intended verb in the X-HTTP-Method-Override header. The override
// middleware is registered first so the request method is rewritten before any
// downstream handler observes it, and a second handler captures req.Method() to
// show the effective verb after the rewrite. The example builds a POST request
// with the default override header set to "delete", then drives it through the
// express application with net/http/httptest. Because the middleware only acts
// on POST requests and upper-cases the override value, the downstream handler
// sees "DELETE", which the example prints deterministically.
func ExampleNew() {
	app := express.New()
	app.Use(methodoverride.New())
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		fmt.Println("effective method:", req.Method())
		res.Send("ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(methodoverride.DefaultHeader, "delete")
	app.ServeHTTP(httptest.NewRecorder(), req)

	// Output: effective method: DELETE
}
