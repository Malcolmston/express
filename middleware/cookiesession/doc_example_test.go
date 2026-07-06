package cookiesession_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/cookiesession"
)

// Example demonstrates a signed, stateless session round-trip with the
// cookiesession middleware. It builds an express application, mounts the
// middleware with a signing secret, and registers two routes: one that writes a
// value into the session and one that reads it back. The first request hits the
// write route; because the session was modified, the middleware serializes it,
// HMAC-signs it, and emits a Set-Cookie header just before the response is
// committed. The example then copies that cookie onto a second request to the
// read route, where the middleware verifies the signature, decodes the payload,
// and makes the stored value available through Get. Finally it prints the value
// recovered on the second request, which is deterministic. A tampered or
// unsigned cookie would fail verification and be ignored, yielding an empty
// session instead.
func Example() {
	app := express.New()
	app.Use(cookiesession.New(cookiesession.Options{Secret: "s3cr3t"}))
	app.Get("/set", func(req *express.Request, res *express.Response, next express.Next) {
		cookiesession.Set(req, "user", "alice")
		res.Send("set")
	})
	app.Get("/get", func(req *express.Request, res *express.Response, next express.Next) {
		v, _ := cookiesession.Get(req, "user")
		res.Send(fmt.Sprint(v))
	})

	// First request stores a value and receives a signed session cookie.
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/set", nil))

	// Second request presents the cookie and reads the value back.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/get", nil)
	for _, c := range rec.Result().Cookies() {
		req2.AddCookie(c)
	}
	app.ServeHTTP(rec2, req2)

	fmt.Println(rec2.Body.String())
	// Output: alice
}
