package bearerauth_test

import (
	"crypto/subtle"
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/bearerauth"
)

// ExampleNew builds bearer authentication middleware and mounts it in front of
// a protected route. The Verify callback compares the presented token against
// the expected value with crypto/subtle.ConstantTimeCompare to avoid leaking
// timing information. The example drives three requests through the application
// with httptest: one with no Authorization header, one with an incorrect
// token, and one with the correct "s3cr3t" token. The first two are
// short-circuited with a 401 and a bare "Bearer" challenge, while the third
// reaches the handler and returns 200, illustrating the challenge-and-reject
// contract of the middleware.
func ExampleNew() {
	app := express.New()
	app.Use(bearerauth.New(bearerauth.Options{
		Verify: func(token string) bool {
			return subtle.ConstantTimeCompare([]byte(token), []byte("s3cr3t")) == 1
		},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("welcome")
	})

	do := func(header string) {
		r := httptest.NewRequest("GET", "/", nil)
		if header != "" {
			r.Header.Set("Authorization", header)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		fmt.Printf("%d %q\n", w.Code, w.Header().Get("WWW-Authenticate"))
	}

	do("")              // missing token
	do("Bearer wrong")  // invalid token
	do("Bearer s3cr3t") // valid token

	// Output:
	// 401 "Bearer"
	// 401 "Bearer"
	// 200 ""
}
