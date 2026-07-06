package version_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/version"
)

// Example mounts the version middleware and issues two requests through
// httptest to show both of its behaviours. A request to the configured version
// path is short-circuited with a JSON body reporting the build, while a request
// to any other route falls through to the ordinary handler yet still carries
// the X-Version response header. Here the version is fixed to "1.2.3" and the
// defaults for path ("/version") and header ("X-Version") are used. Both the
// JSON payload and the header value are deterministic. The Output comment
// therefore pins the exact response for each request.
func Example() {
	app := express.New()
	app.Use(version.New(version.Options{Version: "1.2.3"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("home")
	})

	// The dedicated version endpoint.
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, httptest.NewRequest("GET", "/version", nil))
	fmt.Println("body:", w1.Body.String())

	// Any other route: handler runs, header still present.
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, httptest.NewRequest("GET", "/", nil))
	fmt.Println("home:", w2.Body.String())
	fmt.Println("header:", w2.Header().Get("X-Version"))
	// Output:
	// body: {"version":"1.2.3"}
	// home: home
	// header: 1.2.3
}
