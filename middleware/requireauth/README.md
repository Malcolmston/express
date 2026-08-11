# requireauth — provides middleware that rejects requests unless an authenticated principal has already

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/requireauth.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/requireauth)

Package requireauth provides middleware that rejects requests unless an
authenticated principal has already been placed on the request by an earlier
middleware. It is the express framework's Go analogue of the route guards
found in the Node ecosystem: connect-ensure-login's ensureLoggedIn(),
Passport's req.isAuthenticated() check, and the "is this request logged in?"
gate that most Express apps hand-roll in front of protected routes. Rather
than performing authentication itself, it enforces that some upstream step
succeeded and short-circuits everything that did not.

Reach for this middleware to protect a route, a router, or an entire subtree
once a real authentication middleware runs ahead of it. The typical layout is
a session, JWT, bearer-token, or API-key middleware that, on success, stashes
the resolved user on the request (for example req.Set("user", u)); requireauth
then guarantees that any handler mounted after it only ever runs for an
authenticated caller. Keeping the "who are you" and "you must be someone"
concerns in separate middleware lets you mix and match authentication schemes
while sharing a single, uniform rejection policy.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/requireauth"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./requireauth/`.

```go
app := express.New()

	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		if req.Get("X-Token") == "secret" {
			req.Set("user", "ada")
		}
		next()
	})

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

	call("secret")
	call("")
```

```
200 hello ada
401 Unauthorized
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that responds with 401 unless the configured request value has been set (for example by an authentication middleware). |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the require-auth middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/requireauth).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
