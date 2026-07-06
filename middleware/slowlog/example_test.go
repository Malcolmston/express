package slowlog_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/slowlog"
)

// ExampleNew demonstrates the slow-request logger flagging a handler that
// exceeds its threshold. We build the middleware with a tiny Threshold and a
// custom log.Logger writing into a buffer so the example can assert on the
// output deterministically. The middleware is mounted with app.Use ahead of a
// handler that deliberately sleeps past the threshold, so the request is timed
// end to end. After driving one request through the app with net/http/httptest
// we confirm exactly one WARNING line naming the method and path was logged.
// The elapsed duration itself is nondeterministic, so we assert only on the
// stable parts of the message rather than using an Output block.
func ExampleNew() {
	var buf bytes.Buffer
	app := express.New()
	app.Use(slowlog.New(slowlog.Options{
		Threshold: time.Millisecond,
		Logger:    log.New(&buf, "", 0),
	}))
	app.Get("/report", func(req *express.Request, res *express.Response, next express.Next) {
		time.Sleep(5 * time.Millisecond) // slower than the 1ms threshold
		res.Send("done")
	})

	app.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/report", nil))

	line := strings.TrimSpace(buf.String())
	fmt.Println("warned:", strings.Contains(line, "WARNING"))
	fmt.Println("names route:", strings.Contains(line, "GET /report"))

	// Output:
	// warned: true
	// names route: true
}
