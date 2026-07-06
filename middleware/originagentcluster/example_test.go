package originagentcluster_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/originagentcluster"
)

// ExampleNew shows how to install the Origin-Agent-Cluster middleware on an
// express application. The middleware takes no configuration, so New is called
// with no arguments and registered globally via app.Use before any routes. Once
// installed, every response the app produces carries the header
// Origin-Agent-Cluster: ?1, requesting origin-keyed isolation from the browser.
// The example drives the app with httptest instead of binding a real socket:
// it records a GET request through app.ServeHTTP and then prints the header so
// the deterministic value can be asserted.
func ExampleNew() {
	app := express.New()
	app.Use(originagentcluster.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("hello")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Header().Get("Origin-Agent-Cluster"))
	// Output: ?1
}
