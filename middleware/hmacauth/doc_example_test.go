package hmacauth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/hmacauth"
)

// Example wires the hmacauth middleware into an express application and drives
// it with net/http/httptest. It shares a secret between the signer and the
// middleware, then uses the exported Sign helper to compute the HMAC-SHA256 of
// the request body and places the hex digest in the default X-Signature header.
// Because Sign and New use the identical algorithm and key, the signed request
// authenticates and reaches the handler, which echoes the body to prove the
// middleware restored req.Raw.Body after buffering it. A second request with a
// bogus signature is rejected with 401 before the handler runs. Printing both
// outcomes shows the accept and reject paths side by side.
func Example() {
	secret := []byte("shared-secret")

	app := express.New()
	app.Use(hmacauth.New(hmacauth.Options{Secret: secret}))
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	send := func(body, sig string) int {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		r.Header.Set("X-Signature", sig)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	body := "payload"
	fmt.Println("valid:", send(body, hmacauth.Sign(secret, []byte(body))))
	fmt.Println("invalid:", send(body, "deadbeef"))
	// Output:
	// valid: 200
	// invalid: 401
}
