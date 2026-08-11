# contenttypedefault — provides express middleware that sets a fallback Content-Type on the response when

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/contenttypedefault.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/contenttypedefault)

Package contenttypedefault provides express middleware that sets a fallback
Content-Type on the response when a downstream handler does not set one
itself. It is the Go analogue of the small connect/express helper pattern that
guarantees a default media type (the server-side counterpart of the implicit
Content-Type Express applies via res.type), packaged as a drop-in
express.Handler. It exists so responses never go out with a missing or
accidentally sniffed Content-Type.

Use this middleware when handlers in your application may write raw bytes
without declaring a media type and you want a single, predictable fallback
instead of relying on the client to sniff the body. It is especially useful
for binary or opaque payloads, where the safe default of
"application/octet-stream" tells the client to download rather than render.
Mount it globally with app.Use so every response is covered, or attach it to a
specific router or path prefix to give one subtree its own default.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/contenttypedefault"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./contenttypedefault/`.

```go
app := express.New()
	app.Use(contenttypedefault.New(contenttypedefault.Options{
		Type: "application/octet-stream",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.End()
	})

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	fmt.Println(rr.Header().Get("Content-Type"))
```

```
application/octet-stream
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that, just before the response headers are committed, sets Content-Type to the configured default if the handler has not… |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the middleware. |

### Constants

`DefaultType`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/contenttypedefault).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
