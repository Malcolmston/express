# methodallow — provides middleware that restricts requests to a configured set of HTTP methods

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/methodallow.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/methodallow)

Package methodallow provides middleware that restricts requests to a
configured set of HTTP methods, responding with 405 Method Not Allowed
otherwise. It is the Go equivalent of the small method-guard helpers used in
the Node/Express world (for example connect/express snippets or packages such
as method-not-allowed) that reject unexpected verbs and advertise the allowed
set through an Allow header, implemented here with only the standard library.

Use this middleware when a route or an entire application should accept only a
known subset of HTTP methods and you want unsupported verbs answered with the
correct 405 status and an Allow header instead of falling through to a generic
404 or an unintended handler. It is a good fit for read-only endpoints
(GET/HEAD only), write endpoints that should reject everything but POST, or
any surface where you want to make the permitted verbs explicit and
discoverable to clients and API tooling.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/methodallow"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./methodallow/`.

```go
app := express.New()
	app.Use(methodallow.New(methodallow.Options{Methods: []string{"GET", "POST"}}))
	app.All("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("POST", "/", nil))
	fmt.Printf("POST:   status=%d\n", rec.Code)

	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("DELETE", "/", nil))
	fmt.Printf("DELETE: status=%d allow=%s\n", rec.Code, rec.Header().Get("Allow"))
```

```
POST:   status=200
DELETE: status=405 allow=GET, POST
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that responds with 405 Method Not Allowed, including an Allow header, unless the request method is in the permitted set. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the method-allow middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/methodallow).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
