package poweredby_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/poweredby"
)

// ExampleNew demonstrates setting a custom X-Powered-By response header on
// every response. The middleware is registered with app.Use so it runs ahead
// of the route handler, and here it is configured with an explicit branding
// value via poweredby.Options. A single GET route is added that simply writes
// a body, and the request is driven through the app with httptest so no real
// network listener is required. After ServeHTTP returns, the recorded response
// header carries the configured value, which the example prints to verify the
// deterministic output.
func ExampleNew() {
	app := express.New()
	app.Use(poweredby.New(poweredby.Options{Value: "MyApp/2.0"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	app.ServeHTTP(rr, req)

	fmt.Println(rr.Header().Get("X-Powered-By"))
	// Output: MyApp/2.0
}
