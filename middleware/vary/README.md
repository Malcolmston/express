# vary — provides express middleware that appends fields to the response Vary header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/vary.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/vary)

Package vary provides express middleware that appends fields to the response
Vary header, signalling which request headers a cached response depends on. It
is a port of the widely-used "vary" npm package (the same helper that backs
Express's res.vary() and the CORS middleware's header bookkeeping): given one
or more field names, it ensures each appears exactly once in the response's
Vary header, appending without disturbing entries already set by earlier
middleware.

The Vary header tells caches and CDNs which request headers were used to
select the representation being returned, so a response negotiated on
Accept-Encoding or Origin is not incorrectly served to a client that sent
different values. Use this middleware whenever a response varies by request
header — content negotiation, gzip/br compression, per-Origin CORS
responses, or language selection — and you want the correct Vary field
declared without hand-managing the header string. Declaring Vary is a
correctness concern for shared caches: omitting it can serve the wrong body,
while over-declaring "*" defeats caching entirely.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/vary"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./vary/`.

```go
app := express.New()
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Vary", "Accept-Encoding")
		next()
	})
	app.Use(vary.New(vary.Options{Fields: []string{"accept-encoding", "Origin"}}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	fmt.Println(strings.Join(w.Header().Values("Vary"), ", "))
```

```
Accept-Encoding, Origin
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that appends each configured field to the response's Vary header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the vary middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/vary).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
