package express_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/malcolmston/express"
)

// Example builds a small but complete application and drives one request
// through it without opening a socket. It shows the pieces that make up an
// Express program in Go: express.New returns an *Application whose embedded
// Router exposes Use and the HTTP-method verbs directly, a Handler has the
// signature func(req, res, next), middleware runs before the matched route and
// calls next to continue, and because *Application implements http.Handler the
// whole app can be exercised with net/http/httptest by calling ServeHTTP. The
// route below captures a :name path parameter and echoes it back as JSON.
func Example() {
	app := express.New()

	// Application-wide middleware: annotate every response, then continue.
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-App", "demo")
		next()
	})

	// A route with a named parameter, answering with JSON.
	app.Get("/hello/:name", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON(map[string]string{"hello": req.Params("name")})
	})

	// Drive a request through the app with httptest — no network needed.
	rec := httptest.NewRecorder()
	rec.Body.Reset()
	r := httptest.NewRequest(http.MethodGet, "/hello/world", nil)
	app.ServeHTTP(rec, r)

	fmt.Println("status:", rec.Code)
	fmt.Println("x-app:", rec.Header().Get("X-App"))
	fmt.Println("body:", strings.TrimSpace(rec.Body.String()))
	// Output:
	// status: 200
	// x-app: demo
	// body: {"hello":"world"}
}

// ExampleApplication_Use demonstrates mounting a sub-router at a path prefix.
// A Router created with express.NewRouter is a self-contained stack of
// middleware and routes; mounting it with app.Use("/api", r) strips the prefix
// before the sub-router matches, so the sub-router registers "/status" yet the
// full URL is "/api/status". This is how larger Express apps split routes into
// modular files. The example issues one request and prints the response the
// mounted router produced.
func ExampleApplication_Use() {
	api := express.NewRouter()
	api.Get("/status", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	app := express.New()
	app.Use("/api", api)

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	app.ServeHTTP(rec, r)

	fmt.Println(rec.Code, strings.TrimSpace(rec.Body.String()))
	// Output: 200 ok
}
