# crossoriginembedder — provides express middleware that sets the Cross-Origin-Embedder-Policy (COEP) response

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/crossoriginembedder.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/crossoriginembedder)

Package crossoriginembedder provides express middleware that sets the
Cross-Origin-Embedder-Policy (COEP) response header on every response. It is
the Go analogue of helmet's crossOriginEmbedderPolicy feature (from the Node
helmet package), packaged as a drop-in express.Handler. The one header it
writes, in full, is "Cross-Origin-Embedder-Policy".

COEP controls whether a document is allowed to load cross-origin resources
that have not explicitly opted in to being embedded. Its main purpose is to
help a page become "cross-origin isolated": when a document is served with
both COEP: require-corp (this middleware) and a suitable
Cross-Origin-Opener-Policy, the browser grants access to powerful features
that are otherwise disabled for security reasons, such as SharedArrayBuffer,
high-resolution timers, and precise performance measurement. Reach for this
middleware when you need cross-origin isolation, or simply when you want to
prevent a page from silently embedding cross-origin resources that did not
consent to it.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/crossoriginembedder"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./crossoriginembedder/`.

```go
app := express.New()
	app.Use(crossoriginembedder.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Cross-Origin-Embedder-Policy"))
```

```
require-corp
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the Cross-Origin-Embedder-Policy header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the crossoriginembedder middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/crossoriginembedder).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
