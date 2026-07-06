package nonce_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/nonce"
)

// ExampleNew installs the nonce middleware and demonstrates that a fresh token
// is published for the request. The middleware is registered globally with
// app.Use so it runs before the route handler. Inside the handler the nonce is
// read back with nonce.Nonce(req) and also from res.Locals, showing the two
// publication sites agree. The request is driven in-memory with
// httptest.NewRequest and a recorder via app.ServeHTTP. Because the token is
// random its value is not asserted; instead the example prints only whether a
// non-empty nonce was generated and matched, keeping the output deterministic.
func ExampleNew() {
	app := express.New()
	app.Use(nonce.New(nonce.Options{Bytes: 16}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		fromReq := nonce.Nonce(req)
		fromLocals, _ := res.Locals[nonce.ContextKey].(string)
		fmt.Println(fromReq != "" && fromReq == fromLocals)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	// Output: true
}
