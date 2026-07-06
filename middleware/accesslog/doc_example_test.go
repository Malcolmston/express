package accesslog_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/accesslog"
)

// ExampleNew demonstrates capturing Apache combined-format access logs into a
// buffer. The middleware is mounted first with app.Use so it wraps the response
// writer and observes the final status and byte count, and the handler responds
// with 201 and a short body. After the request completes, one log line is
// written to the configured Options.Writer. Because the log line embeds a
// wall-clock timestamp and the client's ephemeral address, this example does
// not assert the whole line verbatim and instead checks that the request line
// and status were recorded, so no Output block is used.
func ExampleNew() {
	var buf bytes.Buffer

	app := express.New()
	app.Use(accesslog.New(accesslog.Options{Writer: &buf}))
	app.Get("/hello", func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(201).Send("hi there")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))

	if !strings.Contains(buf.String(), "GET /hello") || !strings.Contains(buf.String(), " 201 ") {
		panic("expected a combined-format log line")
	}
}
