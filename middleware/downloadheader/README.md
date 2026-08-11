# downloadheader — provides express middleware that marks responses as file downloads by setting

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/downloadheader.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/downloadheader)

Package downloadheader provides express middleware that marks responses as
file downloads by setting the Content-Disposition response header to
"attachment". It mirrors the behavior of Express's res.attachment() /
res.download() helpers and the Content-Disposition disposition used by the
Node "content-disposition" package, but applies it uniformly to every response
that passes through the middleware rather than at a single call site.

Use this middleware when an entire route tree serves content that browsers
should save to disk instead of rendering inline: report exporters, generated
CSV/PDF endpoints, raw file blobs, or any API whose payloads are meant to be
downloaded. Because it is a plain header-setting handler, mounting it with
app.Use on a subtree is enough to make every response in that subtree a
download; there is no per-response opt-in required.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/downloadheader"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./downloadheader/`.

```go
app := express.New()
	app.Use(downloadheader.New(downloadheader.Options{Filename: "report.csv"}))
	app.Get("/export", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("id,name\n1,alice\n")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/export", nil))

	fmt.Println(rec.Header().Get("Content-Disposition"))
```

```
attachment; filename="report.csv"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets Content-Disposition: attachment so browsers offer to save the response rather than render it. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/downloadheader).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
