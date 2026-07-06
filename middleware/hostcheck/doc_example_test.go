package hostcheck_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/hostcheck"
)

// Example wires the hostcheck middleware into an express application and drives
// it with net/http/httptest. It permits only example.com and its www
// subdomain, leaving Status at its 421 default. Two requests are sent: one with
// an allowed Host header (including a port, which the middleware strips before
// matching) and one with a spoofed host that is rejected before the route
// handler runs. Printing the two status codes shows the allowed request
// reaching the 200 handler while the disallowed request short-circuits with
// 421. This demonstrates how a fixed allowlist neutralizes Host header
// injection.
func Example() {
	app := express.New()
	app.Use(hostcheck.New(hostcheck.Options{
		Allow: []string{"example.com", "www.example.com"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	send := func(host string) int {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Host = host
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	fmt.Println("allowed:", send("www.example.com:8080"))
	fmt.Println("rejected:", send("evil.com"))
	// Output:
	// allowed: 200
	// rejected: 421
}
