package ipblocklist_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/ipblocklist"
)

// Example demonstrates denying a set of clients while leaving the endpoint open
// to everyone else. It constructs the middleware via ipblocklist.New with an
// Options.Block list combining a single banned IP and a banned CIDR range, then
// mounts it globally with app.Use. Two requests are driven through
// net/http/httptest: one whose RemoteAddr lies inside the blocked CIDR and one
// from an unrelated address, with req.RemoteAddr set explicitly because httptest
// defaults every request to 192.0.2.1. Printing the two status codes shows the
// deny decision (403) and the pass-through decision (200). The output is
// deterministic, so an Output block is included.
func Example() {
	app := express.New()
	app.Use(ipblocklist.New(ipblocklist.Options{
		Block: []string{"1.2.3.4", "10.0.0.0/8"},
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

	fmt.Println(call("10.5.5.5"))
	fmt.Println(call("8.8.8.8"))
	// Output:
	// 403
	// 200
}
