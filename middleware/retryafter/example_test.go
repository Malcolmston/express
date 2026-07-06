package retryafter_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/retryafter"
)

// ExampleNew mounts the retry-after middleware and drives two requests through
// httptest to show when the header is and is not attached. The middleware is
// configured to advertise a 30-second back-off and registers an OnBeforeWrite
// hook that runs just before the response headers commit. The first handler
// returns 503 Service Unavailable, so the hook matches the default status set
// and stamps "Retry-After: 30"; the second handler returns a normal 200, so no
// header is added. The example prints the status and Retry-After value for each
// request so the status-driven behavior is visible.
func ExampleNew() {
	app := express.New()
	app.Use(retryafter.New(retryafter.Options{Seconds: 30}))

	app.Get("/down", func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(503).Send("maintenance")
	})
	app.Get("/ok", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("healthy")
	})

	call := func(path string) {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		fmt.Printf("%s -> %d Retry-After=%q\n", path, w.Code, w.Header().Get("Retry-After"))
	}

	call("/down")
	call("/ok")

	// Output:
	// /down -> 503 Retry-After="30"
	// /ok -> 200 Retry-After=""
}
