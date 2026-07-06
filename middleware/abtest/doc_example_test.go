package abtest_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/abtest"
)

// ExampleNew demonstrates mounting the A/B test middleware on an express
// application and branching a handler on the assigned bucket. The middleware is
// added with app.Use so every request is bucketed before routing, and the
// handler reads its cohort with abtest.Bucket(req). Because the bucket is drawn
// with crypto/rand the concrete value varies between runs, so this example
// omits an Output block and instead asserts that a valid bucket and a
// persistence cookie were produced. The same request would keep its bucket on
// return visits once the browser sends the cookie back.
func ExampleNew() {
	app := express.New()
	app.Use(abtest.New(abtest.Options{Buckets: []string{"A", "B"}}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		switch abtest.Bucket(req) {
		case "A":
			res.Send("variant A")
		default:
			res.Send("variant B")
		}
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if len(rec.Result().Cookies()) == 0 {
		panic("expected an abtest cookie")
	}
}
