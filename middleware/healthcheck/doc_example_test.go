package healthcheck_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/healthcheck"
)

// Example mounts the healthcheck middleware on an express application and probes
// it with net/http/httptest. It registers two named checkers via
// healthcheck.Options — both return nil, so both are healthy — and leaves Path
// empty to accept the default "/healthz" endpoint. A GET to that path
// short-circuits the chain: every checker runs and the aggregate result is
// emitted as JSON with a 200 status. Because encoding/json marshals map keys in
// sorted order and both checks pass, the response body is fully deterministic,
// so the Output block asserts the exact status code and JSON payload.
func Example() {
	app := express.New()
	app.Use(healthcheck.New(healthcheck.Options{
		Checkers: map[string]func() error{
			"db":    func() error { return nil },
			"cache": func() error { return nil },
		},
	}))

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	fmt.Println("status:", rec.Code)
	fmt.Println("body:", rec.Body.String())
	// Output:
	// status: 200
	// body: {"status":"ok","checks":{"cache":"ok","db":"ok"}}
}
