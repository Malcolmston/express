# crossoriginresource — provides express middleware that sets the Cross-Origin-Resource-Policy (CORP) response

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/crossoriginresource.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/crossoriginresource)

Package crossoriginresource provides express middleware that sets the
Cross-Origin-Resource-Policy (CORP) response header on every response. It is
the Go analogue of helmet's crossOriginResourcePolicy feature (from the Node
helmet package), packaged as a drop-in express.Handler. The one header it
writes, in full, is "Cross-Origin-Resource-Policy".

CORP lets a server declare which origins are allowed to include (embed, load,
or otherwise consume) the resource it returns. It is the resource-side
counterpart to Cross-Origin-Embedder-Policy: where COEP is a document opting
into strict embedding, CORP is a resource opting into being embedded. Setting
it protects your assets from being loaded by other sites — mitigating
side-channel and speculative-execution attacks such as Spectre by keeping
cross-origin resources out of an attacker's process — and provides the
explicit opt-in that COEP: require-corp pages require of the resources they
embed. Reach for this middleware on endpoints serving images, scripts, fonts,
or APIs that should not be freely embeddable by arbitrary origins.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/crossoriginresource"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./crossoriginresource/`.

```go
app := express.New()
	app.Use(crossoriginresource.New(crossoriginresource.Options{Policy: "cross-origin"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Cross-Origin-Resource-Policy"))
```

```
cross-origin
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the Cross-Origin-Resource-Policy header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the crossoriginresource middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/crossoriginresource).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
