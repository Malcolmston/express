package basicauth_test

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/basicauth"
)

// ExampleNew builds Basic authentication middleware and mounts it in front of a
// protected route. The Verify callback checks the supplied username and
// password with crypto/subtle.ConstantTimeCompare so that credential checks do
// not leak timing information. The example then drives two requests through the
// application with httptest: one carrying no Authorization header, which is
// challenged with a 401 and a WWW-Authenticate response header, and one
// carrying valid "admin:secret" credentials, which reaches the handler and
// returns 200. The printed output shows both outcomes, demonstrating the
// challenge-and-reject contract of the middleware.
func ExampleNew() {
	app := express.New()
	app.Use(basicauth.New(basicauth.Options{
		Realm: "example",
		Verify: func(user, pass string) bool {
			okUser := subtle.ConstantTimeCompare([]byte(user), []byte("admin")) == 1
			okPass := subtle.ConstantTimeCompare([]byte(pass), []byte("secret")) == 1
			return okUser && okPass
		},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("welcome")
	})

	// A request with no credentials is challenged.
	r1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, r1)
	fmt.Printf("no creds: %d %q\n", w1.Code, w1.Header().Get("WWW-Authenticate"))

	// A request with valid credentials reaches the handler.
	cred := base64.StdEncoding.EncodeToString([]byte("admin:secret"))
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.Header.Set("Authorization", "Basic "+cred)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, r2)
	fmt.Printf("valid creds: %d %q\n", w2.Code, w2.Body.String())

	// Output:
	// no creds: 401 "Basic realm=\"example\""
	// valid creds: 200 "welcome"
}
