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

// ExampleApplication_Docs shows how a single call to app.Docs turns an
// application's registered routes into a live OpenAPI 3.1 specification (plus
// Swagger UI, ReDoc, a YAML spec, an AsyncAPI document for event channels and a
// Postman collection). Routes are introspected automatically; Describe adds the
// details introspection cannot infer, and Channel documents socket/event topics.
func ExampleApplication_Docs() {
	app := express.New()

	app.Get("/users/:id", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON(map[string]string{"id": req.Params("id")})
	})

	// Optional: enrich the generated operation.
	app.Describe("GET", "/users/:id", express.RouteDoc{
		Summary: "Fetch a user",
		Tags:    []string{"users"},
	})

	// Optional: document a socket/event channel for the AsyncAPI spec.
	app.Channel("chat.message", express.ChannelDoc{
		Subscribe: &express.MessageDoc{Name: "messageReceived"},
	})

	// Mount /docs, /openapi.json, /openapi.yaml, /redoc, /asyncapi.json and
	// /postman.json.
	app.Docs(express.DocsOptions{Title: "Users API", Version: "1.0.0"})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))

	doc := app.OpenAPI()
	fmt.Println("openapi:", doc.OpenAPI)
	fmt.Println("title:", doc.Info.Title)
	fmt.Println("has /users/{id}:", doc.Paths["/users/{id}"] != nil)
	fmt.Println("served status:", rec.Code)
	// Output:
	// openapi: 3.1.0
	// title: Users API
	// has /users/{id}: true
	// served status: 200
}
