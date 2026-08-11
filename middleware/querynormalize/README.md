# querynormalize — provides middleware that normalizes the request query string

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/querynormalize.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/querynormalize)

Package querynormalize provides middleware that normalizes the request query
string: it lower-cases parameter keys, trims surrounding whitespace from
values, and rebuilds the raw query in a deterministic (key-sorted) order.

It plays the role commonly handled in Node by a query canonicalization step
layered on top of a parser such as qs — the sort of normalization done
before caching, signing, or comparing URLs (comparable in spirit to the
"normalize-url" package's query-key sorting and trimming). By collapsing case,
whitespace, and ordering differences, semantically equivalent requests such as
"?B=2&A=1" and "?a=1&b=2" produce an identical canonical query string, which
improves cache-key stability and makes downstream query handling predictable.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/querynormalize"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./querynormalize/`.

```go
app := express.New()
	app.Use(querynormalize.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(req.Raw.URL.RawQuery)
	})

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/?B=2&A=%20bob%20", nil))

	fmt.Println(rr.Body.String())
```

```
a=bob&b=2
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that rewrites req.Raw.URL.RawQuery in normalized form. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/querynormalize).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
