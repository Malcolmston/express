# etag — provides express middleware that computes a strong ETag for the response body

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/etag.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/etag)

Package etag provides express middleware that computes a strong ETag for the
response body and answers conditional requests with 304 Not Modified. It ports
the response-side ETag behavior that Express enables by default via the Node
"etag" package together with fresh/If-None-Match handling: a validator derived
from the body is emitted, and a matching conditional request is answered with
an empty 304 rather than the full payload.

Use this middleware to cut bandwidth and speed up repeat requests for
responses whose bytes are stable between requests — rendered pages, JSON API
results, static-ish content generated per request. It is most valuable in
front of handlers that regenerate identical output, letting caches and
browsers revalidate cheaply with If-None-Match instead of re-downloading the
body. Mount it with app.Use so it wraps the routes whose responses you want to
tag.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/etag"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./etag/`.

```go
app := express.New()
	app.Use(etag.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("hello")
	})

	rec1 := httptest.NewRecorder()
	app.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/", nil))
	tag := rec1.Header().Get("ETag")

	sum := sha1.Sum([]byte("hello"))
	expected := `"` + hex.EncodeToString(sum[:]) + `"`
	fmt.Println(rec1.Code, tag == expected)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("If-None-Match", tag)
	rec2 := httptest.NewRecorder()
	app.ServeHTTP(rec2, req2)
	fmt.Println(rec2.Code, rec2.Body.Len())
```

```
200 true
304 0
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that sets a strong ETag header (the SHA-1 of the response body, hex-encoded and quoted) on every response with a body. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/etag).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
