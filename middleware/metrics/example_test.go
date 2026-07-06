package metrics_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/metrics"
)

// ExampleNew demonstrates the two-value metrics constructor: New returns both the
// handler to register and the *Metrics accumulator to read from. The handler is
// installed globally with app.Use so it wraps every route and observes the final
// status code. Two routes are added, one returning 200 and one returning 500, and
// each is exercised in-memory with httptest.NewRequest and a recorder through
// app.ServeHTTP. After the requests run, Snapshot reports the accumulated counters
// keyed by "total" and by status class. No Output directive is used because,
// although these counts are stable, metrics values are generally timing- and
// traffic-dependent and should be inspected via Snapshot rather than asserted as
// example output.
func ExampleNew() {
	handler, m := metrics.New()

	app := express.New()
	app.Use(handler)
	app.Get("/ok", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	app.Get("/boom", func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(500).Send("boom")
	})

	for _, path := range []string{"/ok", "/ok", "/boom"} {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	}

	snap := m.Snapshot()
	_ = snap["total"] // total == 3, "2xx" == 2, "5xx" == 1
}
