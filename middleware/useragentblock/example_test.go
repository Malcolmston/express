package useragentblock_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/useragentblock"
)

// Example builds the useragentblock middleware with a deny-list and runs two
// requests through an express application via httptest. The first request uses
// a blocked agent ("curl") and is short-circuited with a 403 before reaching
// the route handler. The second request uses an ordinary browser agent and is
// allowed through to the handler, which returns 200. Matching is
// case-insensitive substring containment, so "CURL/7.0" matches the "curl"
// entry. The printed status codes are deterministic and therefore verified
// with an Output comment.
func Example() {
	app := express.New()
	app.Use(useragentblock.New(useragentblock.Options{
		Block: []string{"curl", "BadBot"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	do := func(ua string) int {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("User-Agent", ua)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	fmt.Println("blocked:", do("CURL/7.0"))
	fmt.Println("allowed:", do("Mozilla/5.0 Firefox/120.0"))
	// Output:
	// blocked: 403
	// allowed: 200
}
