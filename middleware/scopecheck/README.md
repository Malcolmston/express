# scopecheck — provides middleware that authorizes a request only when the caller holds every one

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/scopecheck.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/scopecheck)

Package scopecheck provides middleware that authorizes a request only when the
caller holds every one of a configured set of scopes. It is the express
framework's Go analogue of the OAuth 2.0 / OpenID Connect scope enforcement
that Node services layer on top of express-oauth2-jwt-bearer, express-jwt, or
Passport strategies: an earlier stage validates the access token and exposes
its granted scopes, and this middleware decides whether those scopes cover
what the endpoint demands.

Reach for this middleware to protect an API route whose access token must
carry specific permissions — for example requiring both "read" and "write"
on a mutating endpoint, or "profile" and "email" before returning user data.
The scopes are plain strings compared for equality, so the middleware does not
care whether they arrive from an OAuth token's "scope" claim, an API-key
record, or a session; the Getter callback is the sole bridge between the
request and the scope list.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/scopecheck"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./scopecheck/`.

```go
app := express.New()

	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		req.Set("scopes", []string{"read", "write", "delete"})
		next()
	})

	app.Use(scopecheck.New(scopecheck.Options{
		Required: []string{"read", "write"},
		Getter: func(req *express.Request) []string {
			v, _ := req.Value("scopes")
			scopes, _ := v.([]string)
			return scopes
		},
	}))

	app.Get("/documents", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("granted")
	})

	r := httptest.NewRequest("GET", "/documents", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
```

```
200
granted
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that responds with 403 unless every required scope is present on the request. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the scope-check middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/scopecheck).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
