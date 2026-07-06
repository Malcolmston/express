package crossoriginembedder_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/crossoriginembedder"
)

// Example mounts the Cross-Origin-Embedder-Policy middleware on an express
// application and drives it with net/http/httptest. It uses the zero-value
// configuration, so the middleware emits the default require-corp policy that
// is needed for cross-origin isolation. The route handler simply writes a body,
// showing that the middleware sets its header and then calls next() without
// short-circuiting. After serving the request we print the value of the
// Cross-Origin-Embedder-Policy response header. Because the header setter is
// fully deterministic, the expected output is asserted below.
func Example() {
	app := express.New()
	app.Use(crossoriginembedder.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Cross-Origin-Embedder-Policy"))
	// Output:
	// require-corp
}
