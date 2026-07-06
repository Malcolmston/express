package querylimit_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/querylimit"
)

// ExampleNew demonstrates rejecting requests whose raw query string is too
// long. The middleware is registered with app.Use and configured through
// querylimit.Options with a small MaxLength so the limit is easy to exceed. A
// GET route is added that would return "ok" for accepted requests. Two
// requests are driven through the app with httptest: one whose query fits
// within the limit and one whose query overflows it. The example prints the
// resulting status codes, showing the within-limit request succeeding with 200
// and the oversized request being short-circuited with a 414 URI Too Long
// response before the handler runs.
func ExampleNew() {
	app := express.New()
	app.Use(querylimit.New(querylimit.Options{MaxLength: 10}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	small := httptest.NewRecorder()
	app.ServeHTTP(small, httptest.NewRequest("GET", "/?a=1", nil))

	big := httptest.NewRecorder()
	app.ServeHTTP(big, httptest.NewRequest("GET", "/?q=aaaaaaaaaaaaaaaaaaaa", nil))

	fmt.Println(small.Code)
	fmt.Println(big.Code)
	// Output:
	// 200
	// 414
}
