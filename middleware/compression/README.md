# compression — provides express middleware that gzip-compresses response bodies for clients

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/compression.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/compression)

Package compression provides express middleware that gzip-compresses response
bodies for clients that advertise gzip support via the Accept-Encoding request
header. It is the Go analogue of the Node compression middleware
(expressjs/compression), packaged as a drop-in express.Handler, and it trades
a small amount of CPU for a large reduction in bytes on the wire for
compressible payloads such as HTML, JSON, CSS, and JavaScript.

Use this middleware when you serve text-based responses to bandwidth-limited
or latency-sensitive clients and want smaller transfers without touching your
handlers. Mount it once near the top of the chain with app.Use so that it
wraps every downstream response, or attach it to a specific router or path
prefix to compress only part of the tree. Because it buffers the body to make
a size-based decision, it is best suited to ordinary buffered responses rather
than long-lived streams; place any handler that needs raw, unbuffered access
to the response writer ahead of it.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/compression"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./compression/`.

```go
app := express.New()
	app.Use(compression.New(compression.Options{MinLength: 16}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(strings.Repeat("hello world ", 64))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	fmt.Println(rr.Header().Get("Content-Encoding"))
```

```
gzip
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that gzip-compresses eligible responses. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the compression middleware. |

### Constants

`DefaultMinLength`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/compression).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
