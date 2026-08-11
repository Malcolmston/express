# basepath — provides middleware that strips a fixed prefix from the request path

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/basepath.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/basepath)

Package basepath provides middleware that strips a fixed prefix from the
request path, allowing an application mounted under a sub-path (for example
behind a reverse proxy at "/app") to be written as though it were served from
root. It is the Go analogue of Node helpers such as express-basepath and the
mounting behavior of Express sub-apps, packaged as a drop-in express.Handler
that rewrites the router's match path before downstream routes run.

Use this middleware when the public URL space and the application's internal
route definitions differ by a constant prefix. A common case is a service
deployed under "/app" by an upstream proxy or gateway while its handlers are
authored against clean paths like "/users" and "/health". Instead of prefixing
every route registration, mount basepath first and let each request arrive at
the routes with the prefix already removed, keeping route definitions portable
across mount points.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/basepath"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./basepath/`.

```go
app := express.New()
	app.Use(basepath.New(basepath.Options{Prefix: "/app"}))
	app.Get("/users", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("matched " + req.Path())
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/app/users", nil))

	fmt.Println(rec.Body.String())
```

```
matched /users
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that strips Options.Prefix from req.Raw.URL.Path. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the basepath middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/basepath).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
