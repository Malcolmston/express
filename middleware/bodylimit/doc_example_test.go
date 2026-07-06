package bodylimit_test

import (
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/bodylimit"
)

// ExampleNew mounts the body-limit middleware with a tiny 8-byte ceiling in
// front of an echo handler. The example drives two POST requests through the
// application with httptest: one whose body fits within the limit, which
// reaches the handler and echoes back with 200, and one whose declared
// Content-Length exceeds the limit, which is rejected immediately with 413
// Payload Too Large before the handler runs. The printed output shows both
// status codes, demonstrating how the Content-Length gate refuses oversized
// uploads up front. A MaxBytes of zero or less would disable the limit
// entirely.
func ExampleNew() {
	app := express.New()
	app.Use(bodylimit.New(bodylimit.Options{MaxBytes: 8}))
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("stored")
	})

	do := func(body string) {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		fmt.Printf("%d %q\n", w.Code, w.Body.String())
	}

	do("tiny")                 // 4 bytes: within limit
	do("way too much payload") // over 8 bytes: rejected

	// Output:
	// 200 "stored"
	// 413 "Request Entity Too Large"
}
