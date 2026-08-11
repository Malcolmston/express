# nosniff — provides middleware that sets the response header X-Content-Type-Options

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/nosniff.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/nosniff)

Package nosniff provides middleware that sets the response header
X-Content-Type-Options: nosniff on every request. It is the express port of
Helmet's noSniff middleware (and the equivalent dontSniffMimetype behavior of
the Node lusca and koa-helmet families), which in turn wraps the single header
that all modern browsers honor to disable content-type sniffing.

Use it whenever you serve responses whose bytes could be misinterpreted if a
browser ignored the declared Content-Type: user-uploaded files, JSON or text
endpoints that an attacker might coax a browser into executing as HTML or
JavaScript, and any response where "the server said so" should be the final
word on the media type. It is cheap enough to enable globally and is part of
most sensible security baselines, which is why Helmet turns it on by default.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/nosniff"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./nosniff/`.

```go
app := express.New()
	app.Use(nosniff.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("X-Content-Type-Options"))
```

```
nosniff
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that sets X-Content-Type-Options: nosniff. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/nosniff).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
