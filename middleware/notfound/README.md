# notfound — provides a terminal handler that responds with 404 Not Found for requests that fall

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/notfound.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/notfound)

Package notfound provides a terminal handler that responds with 404 Not Found
for requests that fall through the router unmatched. It plays the role of
Express's built-in final 404 handler and of common Node snippets like the
"catch-all" app.use((req, res) => res.status(404)...) pattern and the
express-error-handler style notFound helper, giving unmatched routes one
consistent response instead of the framework default.

Use it as the last entry you register, after every route and every other piece
of middleware, so it only runs when nothing earlier produced a response. This
is the idiomatic way to control the body, content type, and shape of your "not
found" answer application-wide: a plain-text message for humans, or a JSON
error object for API clients that expect machine-readable errors.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/notfound"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./notfound/`.

```go
app := express.New()
	app.Get("/home", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("home")
	})
	app.Use(notfound.New())

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
```

```
404
Not Found
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns a terminal handler that always responds 404. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the notfound handler. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/notfound).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
