package latencyheader_test

import (
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/latencyheader"
)

// ExampleNew demonstrates wiring the latency-header middleware into an express
// application and driving it with net/http/httptest. The middleware is
// registered first so that its OnBeforeWrite hook wraps the entire handler
// chain, and a controllable clock is injected through Options.Now so the
// measured interval is deterministic instead of depending on real wall-clock
// time. The route handler advances the fake clock by 25ms before sending its
// body, which is exactly the elapsed value the middleware stamps into the
// response header. After ServeHTTP runs the recorder captures the committed
// response, and the example prints the X-Response-Latency header to show the
// recorded processing time in whole milliseconds.
func ExampleNew() {
	cur := time.Unix(0, 0)
	app := express.New()
	app.Use(latencyheader.New(latencyheader.Options{
		Now: func() time.Time { return cur },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		cur = cur.Add(25 * time.Millisecond)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	fmt.Println(rec.Header().Get("X-Response-Latency"))
	// Output: 25
}
