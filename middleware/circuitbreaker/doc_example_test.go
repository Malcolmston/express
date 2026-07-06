package circuitbreaker_test

import (
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/circuitbreaker"
)

// ExampleNew mounts the circuit-breaker middleware in front of a flaky handler
// and shows the breaker opening after repeated failures. The example injects a
// fixed clock through Options.Now so the time-based behaviour is fully
// deterministic. With a threshold of 2, the first two requests reach the
// handler and return its 500 response; the breaker then opens, so the third
// request is short-circuited with a 503 carrying the configured message without
// ever invoking the handler. Advancing the injected clock past the cooldown
// moves the breaker to half-open, and because the dependency has recovered the
// trial request succeeds with 200 and the circuit closes again. The output
// traces this closed to open to recovered progression.
func ExampleNew() {
	now := time.Unix(0, 0)
	healthy := false

	app := express.New()
	app.Use(circuitbreaker.New(circuitbreaker.Options{
		Threshold: 2,
		Cooldown:  30 * time.Second,
		Message:   "circuit open",
		Now:       func() time.Time { return now },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		if healthy {
			res.Send("ok")
			return
		}
		res.Status(500).Send("upstream down")
	})

	do := func() {
		r := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		fmt.Printf("%d %q\n", w.Code, w.Body.String())
	}

	do() // failure 1 (handler 500)
	do() // failure 2 (handler 500) -> trips open
	do() // short-circuited by open circuit

	// Recover: advance past the cooldown and let the dependency succeed.
	now = now.Add(31 * time.Second)
	healthy = true
	do() // half-open trial succeeds -> circuit closes

	// Output:
	// 500 "upstream down"
	// 500 "upstream down"
	// 503 "circuit open"
	// 200 "ok"
}
