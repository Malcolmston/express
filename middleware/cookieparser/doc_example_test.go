package cookieparser_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/cookieparser"
)

// Example demonstrates reading request cookies with the cookieparser
// middleware. It builds an express application and mounts the middleware, which
// parses every cookie on the incoming request into a map[string]string and
// stores it for downstream use. The handler retrieves that map with the From
// helper and looks up individual cookies by name. The driving request carries
// two cookies, one with a plain value and one whose value contains a space that
// was URL-encoded on the wire; the middleware URL-unescapes values so the space
// is restored. The example prints both decoded values, which is fully
// deterministic. Note that From always returns a non-nil map, so handlers can
// index it safely even when no cookies were sent.
func Example() {
	app := express.New()
	app.Use(cookieparser.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		cookies := cookieparser.From(req)
		fmt.Println(cookies["session"])
		fmt.Println(cookies["greeting"])
		res.Send("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	req.AddCookie(&http.Cookie{Name: "greeting", Value: "hello%20world"})
	app.ServeHTTP(httptest.NewRecorder(), req)

	// Output:
	// abc123
	// hello world
}
