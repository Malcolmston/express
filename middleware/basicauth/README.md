# basicauth — provides HTTP Basic authentication middleware for the express framework

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/basicauth.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/basicauth)

Package basicauth provides HTTP Basic authentication middleware for the
express framework. It challenges unauthenticated requests with a
WWW-Authenticate header and rejects invalid credentials with 401. It is the Go
analogue of Node middleware such as express-basic-auth or the classic
basic-auth helper, packaged as a drop-in express.Handler.

Use this middleware to protect an entire application or a subtree of routes
behind a username/password gate without standing up sessions, cookies, or a
login page. Because the browser's built-in credential dialog is triggered by
the challenge, Basic auth is a good fit for internal tools, staging
environments, health dashboards, and machine-to-machine calls where a full
authentication flow would be overkill. Mount it with app.Use for a global
guard, or attach it to a specific router or path prefix to protect only part
of the tree.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/basicauth"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./basicauth/`.

```go
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

	r1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, r1)
	fmt.Printf("no creds: %d %q\n", w1.Code, w1.Header().Get("WWW-Authenticate"))

	cred := base64.StdEncoding.EncodeToString([]byte("admin:secret"))
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.Header.Set("Authorization", "Basic "+cred)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, r2)
	fmt.Printf("valid creds: %d %q\n", w2.Code, w2.Body.String())
```

```
no creds: 401 "Basic realm=\"example\""
valid creds: 200 "welcome"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that enforces HTTP Basic authentication. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the Basic authentication middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/basicauth).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
