package correlationid_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/correlationid"
)

// Example wires the correlation-id middleware into an express application and
// drives it with net/http/httptest. It configures a fixed Generator so the
// output is deterministic, but in production you would leave Generator nil to
// get a random 32-character hex id. The route handler reads the id back from
// the request via req.Value(correlationid.ContextKey), demonstrating how
// downstream handlers recover the value for logging. After serving the request
// we inspect the echoed X-Correlation-Id response header, which carries the
// same id the middleware assigned. Because no inbound header is present, the
// middleware generates a new id rather than preserving one.
func Example() {
	app := express.New()
	app.Use(correlationid.New(correlationid.Options{
		Generator: func() string { return "fixed-id-123" },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		id, _ := req.Value(correlationid.ContextKey)
		fmt.Println("handler saw:", id)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println("response header:", rec.Header().Get(correlationid.DefaultHeader))
	// Output:
	// handler saw: fixed-id-123
	// response header: fixed-id-123
}
