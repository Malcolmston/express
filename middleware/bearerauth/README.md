# bearerauth — provides middleware that authenticates requests using an opaque bearer token supplied

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/bearerauth.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/bearerauth)

Package bearerauth provides middleware that authenticates requests using an
opaque bearer token supplied in the Authorization header. It is the express
framework's Go analogue of the "Bearer" scheme handled by Node middleware such
as express-bearer-token and passport-http-bearer, packaged as a drop-in
express.Handler that challenges callers with a WWW-Authenticate header and
rejects anything without a valid token.

Reach for this middleware when you want to gate routes behind a token instead
of a username/password pair: API endpoints consumed by scripts and services,
mobile or SPA back ends that already mint access tokens, webhook receivers, or
any machine-to-machine call that carries a credential in the standard
"Authorization: Bearer <token>" form. Because the token is opaque to the
middleware, it works equally well with random session identifiers, API keys,
or signed tokens such as JWTs; the meaning of the token is entirely defined by
the caller-supplied Verify callback.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/bearerauth"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./bearerauth/`.

```go
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

	do("")
	do("Bearer wrong")
	do("Bearer s3cr3t")
```

```
401 "Bearer"
401 "Bearer"
200 ""
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that reads an "Authorization: Bearer <token>" header and rejects the request with 401 when the token is missing or invalid. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the bearer authentication middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/bearerauth).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
