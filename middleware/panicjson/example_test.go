package panicjson_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/panicjson"
)

// ExampleNew demonstrates using the panic-recovery middleware as an outer safety
// net around a handler that fails. The middleware is registered first via
// app.Use so its deferred recover wraps every downstream handler in the chain.
// The route below deliberately panics, simulating a nil dereference or failed
// assertion, which would otherwise unwind the request goroutine. Instead the
// middleware catches the panic, logs it (here to a discarded logger for
// determinism), and, because nothing has been written yet, sends a fixed 500
// JSON error. The example drives the request through app.ServeHTTP with an
// httptest recorder and prints the status code and body to show the recovered
// response.
func ExampleNew() {
	app := express.New()
	app.Use(panicjson.New(panicjson.Options{
		Logger: log.New(io.Discard, "", 0),
	}))
	app.Get("/boom", func(req *express.Request, res *express.Response, next express.Next) {
		panic("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 500
	// {"error":"Internal Server Error"}
}
