package expires_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/expires"
)

// Example demonstrates stamping responses with an Expires header a fixed
// duration into the future. It builds an express app and mounts expires.New
// with an Options value giving a one-hour lifetime, then registers a route and
// drives a request through the app with net/http/httptest. The emitted Expires
// header is parsed back with http.ParseTime and compared against the request
// time to confirm it lies in the future. No Output comment is used because the
// header carries the wall-clock moment of the request and is therefore not
// deterministic; the example instead prints only a stable boolean.
func Example() {
	app := express.New()
	app.Use(expires.New(expires.Options{Duration: time.Hour}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	exp, err := http.ParseTime(rec.Header().Get("Expires"))
	fmt.Println(err == nil && exp.After(time.Now()))
	// Output: true
}
