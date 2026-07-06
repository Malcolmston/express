package ratelimit_test

import (
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/ratelimit"
)

// ExampleNew demonstrates the fixed-window rate limiter rejecting a client that
// exceeds its allowance. The middleware is registered with app.Use and
// configured through ratelimit.Options with a Max of 1 request per Window; a
// fixed Now clock is injected so the window boundary is deterministic and the
// output is stable. A GET route is added that returns "ok" for admitted
// requests. Two requests are driven from the same RemoteAddr through the app
// with httptest so both hash to the same IP bucket. The example prints the two
// status codes, showing the first request admitted with 200 and the second
// short-circuited with 429 Too Many Requests because it exhausted the window's
// single-request budget.
func ExampleNew() {
	clock := time.Unix(1000, 0)
	app := express.New()
	app.Use(ratelimit.New(ratelimit.Options{
		Max:    1,
		Window: time.Minute,
		Now:    func() time.Time { return clock },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	do := func() int {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "1.2.3.4:5555"
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	fmt.Println(do())
	fmt.Println(do())
	// Output:
	// 200
	// 429
}
