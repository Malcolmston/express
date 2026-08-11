# bodylimit — provides express middleware that rejects requests whose body exceeds a configured size

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/bodylimit.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/bodylimit)

Package bodylimit provides express middleware that rejects requests whose body
exceeds a configured size, guarding handlers against oversized uploads. It is
the Go analogue of the size limit enforced by Node's body-parser and the
raw-body module (the "limit" option surfaced by express.json,
express.urlencoded, and express.raw), extracted into a standalone
express.Handler that can protect any route regardless of content type.

Use this middleware to defend an application against memory-exhaustion and
denial-of-service attempts in which a client streams an unbounded request
body. Because it caps the number of bytes any downstream handler can read, it
is valuable in front of file-upload endpoints, JSON APIs, and proxy routes,
and is cheap enough to mount globally with app.Use so that every route
inherits a sane ceiling. Set a smaller per-route limit on endpoints that
should only ever receive tiny payloads.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/bodylimit"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./bodylimit/`.

```go
app := express.New()
	app.Use(bodylimit.New(bodylimit.Options{MaxBytes: 8}))
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("stored")
	})

	do := func(body string) {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		fmt.Printf("%d %q\n", w.Code, w.Body.String())
	}

	do("tiny")
	do("way too much payload")
```

```
200 "stored"
413 "Request Entity Too Large"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware enforcing a maximum request body size. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the bodylimit middleware. |

### Constants

`DefaultMaxBytes`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/bodylimit).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
