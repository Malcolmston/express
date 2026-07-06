package frameguard_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/frameguard"
)

// Example mounts the frameguard middleware on an express application and drives
// it with net/http/httptest. It configures Options{Action: "deny"} to show that
// the directive is matched case-insensitively and normalized to the canonical
// "DENY" value. The middleware sets the X-Frame-Options response header on every
// response before the route handler runs, without touching the body or
// short-circuiting the chain. After serving a request we read the header back
// off the recorder to confirm the directive that was sent. The value is
// deterministic, so the Output block asserts it exactly.
func Example() {
	app := express.New()
	app.Use(frameguard.New(frameguard.Options{Action: "deny"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println("X-Frame-Options:", rec.Header().Get("X-Frame-Options"))
	// Output:
	// X-Frame-Options: DENY
}
