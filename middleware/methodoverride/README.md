# methodoverride — provides middleware that overrides the HTTP request method using either a request

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/methodoverride.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/methodoverride)

Package methodoverride provides middleware that overrides the HTTP request
method using either a request header or a query-string parameter. This lets
HTML forms (which can only issue GET and POST) emulate verbs such as PUT,
PATCH, and DELETE. It is the Go port of the Node method-override middleware
(the method-override npm package used with connect/express), reproducing its
core header- and query-based override behavior with only the standard library.

Use this middleware in front of a route table that is built around REST verbs
but must be reachable from plain HTML forms or clients that can only send
POST. By letting a form carry the intended verb in a hidden _method field or
in the X-HTTP-Method-Override header, you can keep RESTful routes (app.Put,
app.Delete, app.Patch) while the browser still submits an ordinary POST. It is
equally useful for constrained HTTP clients or proxies that strip non-standard
verbs.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/methodoverride"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./methodoverride/`.

```go
app := express.New()
	app.Use(methodoverride.New())
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		fmt.Println("effective method:", req.Method())
		res.Send("ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(methodoverride.DefaultHeader, "delete")
	app.ServeHTTP(httptest.NewRecorder(), req)
```

```
effective method: DELETE
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that rewrites req.Raw.Method from the configured header or query parameter. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the method-override middleware. |

### Constants

`DefaultHeader`, `DefaultQuery`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/methodoverride).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
