package concurrencylimit_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/concurrencylimit"
)

// Example demonstrates capping in-flight requests with the concurrencylimit
// middleware. It builds an express application and mounts a limiter that allows
// at most two requests to run concurrently, rejecting any overflow with a 503.
// A single request is then driven through with net/http/httptest. Because that
// lone request takes one of the two available slots and releases it as soon as
// the handler returns, it is admitted and served normally. The example prints
// the resulting status code, which is deterministic for a request that fits
// comfortably under the concurrency ceiling. Under real load, once Max
// simultaneous requests are being processed, further requests would instead
// receive the 503 Service Unavailable response configured via Options.Message.
func Example() {
	app := express.New()
	app.Use(concurrencylimit.New(concurrencylimit.Options{
		Max:     2,
		Message: "Busy, try again later",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	fmt.Println(rr.Code)
	// Output: 200
}
