package ipallowlist_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/ipallowlist"
)

// Example demonstrates gating an application so that only clients from an
// approved exact address or CIDR range may reach the route, while everyone else
// receives 403 Forbidden. It constructs the middleware via ipallowlist.New with
// an Options.Allow list mixing a single IP and a CIDR block, then mounts it
// globally with app.Use. Two requests are driven through net/http/httptest: one
// whose RemoteAddr falls inside the allowed CIDR and one that does not, with
// req.RemoteAddr set explicitly because httptest defaults every request to
// 192.0.2.1. Printing the two status codes shows the allow decision (200) and
// the deny decision (403). The output is deterministic, so an Output block is
// included.
func Example() {
	app := express.New()
	app.Use(ipallowlist.New(ipallowlist.Options{
		Allow: []string{"10.0.0.5", "192.168.0.0/24"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	call := func(clientIP string) int {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = clientIP + ":1234"
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, r)
		return rec.Code
	}

	fmt.Println(call("192.168.0.42"))
	fmt.Println(call("8.8.8.8"))
	// Output:
	// 200
	// 403
}
