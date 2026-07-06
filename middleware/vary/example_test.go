package vary_test

import (
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/vary"
)

// Example demonstrates the vary middleware declaring the request headers a
// response depends on. An earlier handler pre-sets Vary to "Accept-Encoding",
// then the vary middleware is configured to add both "Accept-Encoding" and
// "Origin". Because de-duplication is case-insensitive and token-aware, the
// already-present "Accept-Encoding" is not repeated while "Origin" is
// appended. The request is driven through httptest and the resulting Vary
// header values are printed. The outcome is deterministic, so the result is
// checked with an Output comment.
func Example() {
	app := express.New()
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Vary", "Accept-Encoding")
		next()
	})
	app.Use(vary.New(vary.Options{Fields: []string{"accept-encoding", "Origin"}}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	fmt.Println(strings.Join(w.Header().Values("Vary"), ", "))
	// Output: Accept-Encoding, Origin
}
