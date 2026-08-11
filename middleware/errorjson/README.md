# errorjson — provides an express error-handling middleware that renders unhandled errors as a JSON

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/errorjson.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/errorjson)

Package errorjson provides an express error-handling middleware that renders
unhandled errors as a JSON object of the form {"error": "<message>"}. It plays
the role of a terminal JSON error handler, comparable to the default error
responder in Express plus a small custom middleware such as the ones commonly
written on top of the Node "http-errors" package: instead of Express's HTML
error page, every error that propagates to the end of the chain is serialized
as a compact JSON document suited to API clients.

Use this handler when the application is an API (or any client that expects
JSON) and you want a single, consistent error envelope regardless of where an
error originates. Register it last, after all routes and other middleware, so
that it acts as the catch-all: express dispatches to error handlers only once
a handler calls next(err), so anything that forwards an error eventually lands
here.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/errorjson"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./errorjson/`.

```go
app := express.New()
	app.Get("/boom", func(req *express.Request, res *express.Response, next express.Next) {
		next(errors.New("something failed"))
	})
	app.Use(errorjson.New())

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
```

```
500
{"error":"something failed"}
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.ErrorHandler` | New returns an express.ErrorHandler that writes the error's message as a JSON document. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the JSON error handler. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/errorjson).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
