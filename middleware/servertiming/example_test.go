package servertiming_test

import (
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/servertiming"
)

// ExampleNew wires the server-timing middleware into an express application and
// drives it with net/http/httptest. The middleware is registered first so that
// its OnBeforeWrite hook wraps the whole handler chain and emits the accumulated
// Server-Timing header just before the response is committed. Inside the handler,
// From retrieves the per-request Metrics collector and records two measurements:
// a bare "db" timing and a "render" timing carrying a human-readable description.
// After ServeHTTP runs, the recorder holds the response whose Server-Timing header
// lists both entries in name;desc="...";dur=<ms> form. The header value depends on
// the recorded durations and is therefore not asserted with an Output comment; the
// example instead prints whether the header was populated, which is always true
// once at least one metric is added.
func ExampleNew() {
	app := express.New()
	app.Use(servertiming.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		m := servertiming.From(req)
		m.Add("db", 12500*time.Microsecond)
		m.AddWithDesc("render", "template render", 3*time.Millisecond)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	fmt.Println(rec.Header().Get(servertiming.HeaderName) != "")
}
