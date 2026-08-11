# apikey — provides middleware that authenticates requests using an API key supplied via a request

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/apikey.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/apikey)

Package apikey provides middleware that authenticates requests using an API
key supplied via a request header or query-string parameter. It is the Go
analogue of the many Node API-key gatekeepers (for example
express-api-key-auth style middleware) and complements the package's basicauth
middleware, packaged as a drop-in express.Handler that rejects unauthenticated
requests with 401.

Use this middleware to protect an application or a subtree of routes behind a
shared secret without sessions, cookies, or a login page. It fits internal
services, server-to-server calls, and public APIs that issue per-client keys,
where a lightweight bearer-style credential is more appropriate than
interactive authentication. Mount it with app.Use for a global guard or attach
it to a specific router to protect only part of the tree.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/apikey"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./apikey/`.

```go
app := express.New()
	app.Use(apikey.New(apikey.Options{
		Query: "api_key",
		Keys:  []string{"key1", "key2"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	authed := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-API-Key", "key1")
	app.ServeHTTP(authed, r)

	anon := httptest.NewRecorder()
	app.ServeHTTP(anon, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(authed.Code, anon.Code)
```

```
200 401
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that requires a valid API key. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the API-key middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/apikey).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
