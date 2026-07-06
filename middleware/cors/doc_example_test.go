package cors_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/cors"
)

// Example configures the CORS middleware with an explicit allow-list and drives
// it with net/http/httptest. It first issues a normal GET request carrying an
// allowed Origin and shows that the route handler runs and the
// Access-Control-Allow-Origin header is set to the concrete origin. It then
// issues an OPTIONS preflight request and shows that the middleware
// short-circuits with 204 No Content and advertises the permitted methods
// without invoking the route handler. Everything here is deterministic, so the
// expected output is asserted below.
func Example() {
	app := express.New()
	app.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"https://app.example.com"},
		AllowedMethods: []string{"GET", "POST"},
		MaxAge:         600,
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	// A simple cross-origin GET from an allowed origin.
	get := httptest.NewRecorder()
	greq := httptest.NewRequest(http.MethodGet, "/", nil)
	greq.Header.Set("Origin", "https://app.example.com")
	app.ServeHTTP(get, greq)
	fmt.Println("GET status:", get.Code)
	fmt.Println("GET allow-origin:", get.Header().Get("Access-Control-Allow-Origin"))
	fmt.Println("GET body:", get.Body.String())

	// A preflight OPTIONS request is answered with 204 and no handler run.
	pre := httptest.NewRecorder()
	preq := httptest.NewRequest(http.MethodOptions, "/", nil)
	preq.Header.Set("Origin", "https://app.example.com")
	app.ServeHTTP(pre, preq)
	fmt.Println("preflight status:", pre.Code)
	fmt.Println("preflight allow-methods:", pre.Header().Get("Access-Control-Allow-Methods"))

	// Output:
	// GET status: 200
	// GET allow-origin: https://app.example.com
	// GET body: ok
	// preflight status: 204
	// preflight allow-methods: GET, POST
}
