package ienoopen_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/ienoopen"
)

// Example demonstrates mounting the ienoopen middleware globally so that every
// response carries the X-Download-Options: noopen header. It builds an Express
// application, registers ienoopen.New() with app.Use, and adds a single route
// that serves a download-like payload. The request is then driven in-process
// with net/http/httptest, without binding a socket, so the example is fully
// self-contained. After ServeHTTP returns, the recorder exposes the response
// headers, letting us confirm the hardening header was applied. The Output
// block is deterministic because the header value is a fixed constant.
func Example() {
	app := express.New()
	app.Use(ienoopen.New())
	app.Get("/download", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("file-contents")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/download", nil))

	fmt.Println(rec.Header().Get("X-Download-Options"))
	// Output: noopen
}
