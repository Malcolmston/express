# decompress — provides express middleware that transparently decompresses gzip-encoded request bodies

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/decompress.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/decompress)

Package decompress provides express middleware that transparently decompresses
gzip-encoded request bodies so downstream handlers read plain data. It is the
inverse of response compression: where a compression middleware shrinks
outgoing responses, decompress inflates incoming requests, mirroring the
request-body inflation that Node's HTTP stack and middleware such as
body-parser perform when a client sends a gzip-encoded payload.

Use this middleware when clients — mobile apps, IoT devices, or services
uploading large JSON, logs, or telemetry — send request bodies with
"Content-Encoding: gzip" to save bandwidth. Without it, a handler reading
req.Raw.Body would see raw compressed bytes; with it, the handler reads the
original plain data as though no encoding had been applied. Mount it with
app.Use ahead of any body-parsing middleware or route handler that consumes
the request body, so the body is already inflated by the time they read it.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/decompress"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./decompress/`.

```go
app := express.New()
	app.Use(decompress.New())
	app.Post("/upload", func(req *express.Request, res *express.Response, next express.Next) {
		data, _ := io.ReadAll(req.Raw.Body)
		res.Send(string(data))
	})

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte("hello compressed world"))
	gw.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
```

```
hello compressed world
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that, when the request's Content-Encoding is gzip, wraps req.Raw.Body in a gzip reader. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/decompress).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
