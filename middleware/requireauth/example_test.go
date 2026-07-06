package requireauth_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/requireauth"
)

// ExampleNew wires requireauth into an express app behind a toy authentication
// middleware and drives two requests through it with httptest. The first
// request carries a valid "token" header, so the upstream middleware resolves a
// user and stores it with req.Set("user", ...); requireauth sees a non-nil
// value and lets the protected handler run. The second request omits the token,
// the upstream middleware sets nothing, and requireauth short-circuits with a
// 401 Unauthorized before the handler is reached. The example prints each
// status and body so the guard's allow-and-reject behavior is visible.
func ExampleNew() {
	app := express.New()

	// Pretend authentication: a valid token resolves to a user.
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		if req.Get("X-Token") == "secret" {
			req.Set("user", "ada")
		}
		next()
	})

	// Require that some earlier middleware populated the "user" value.
	app.Use(requireauth.New(requireauth.Options{}))

	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		user, _ := req.Value("user")
		res.Send(fmt.Sprintf("hello %v", user))
	})

	call := func(token string) {
		r := httptest.NewRequest("GET", "/", nil)
		if token != "" {
			r.Header.Set("X-Token", token)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		fmt.Printf("%d %s\n", w.Code, w.Body.String())
	}

	call("secret") // authenticated
	call("")       // missing credentials

	// Output:
	// 200 hello ada
	// 401 Unauthorized
}
