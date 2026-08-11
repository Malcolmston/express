# pagination — provides middleware that parses the common "page" and "limit" query-string parameters

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/pagination.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/pagination)

Package pagination provides middleware that parses the common "page" and
"limit" query-string parameters into a normalized, bounds-checked Pagination
value stored on the request. It ports the ubiquitous limit/offset paging
convention used by Node helpers such as express-paginate and the manual
req.query.page / req.query.limit parsing found in most Express APIs, turning
that repetitive boilerplate into a single reusable handler.

Use it on any list endpoint that returns results in pages. Instead of each
route re-reading and re-validating the query parameters, the middleware does
the parsing once and exposes a ready-to-use Pagination struct carrying the
requested Page, the effective Limit, and a precomputed Offset suitable for
passing straight to a database LIMIT/OFFSET clause or slice bounds.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/pagination"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./pagination/`.

```go
app := express.New()
	app.Use(pagination.New(pagination.Options{DefaultLimit: 20, MaxLimit: 50}))
	app.Get("/items", func(req *express.Request, res *express.Response, next express.Next) {
		p := pagination.From(req)
		res.JSON(map[string]int{
			"page":   p.Page,
			"limit":  p.Limit,
			"offset": p.Offset,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/items?page=3&limit=999", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
```

```
{"limit":50,"offset":100,"page":3}
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that reads ?page and ?limit, clamps them to sensible bounds, and stores the resulting Pagination on the request. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the pagination middleware. |
| `Pagination` | Pagination holds the parsed and normalized paging parameters for a request. |

<details>
<summary><code>Pagination</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func From(req *express.Request) Pagination` | From returns the Pagination stored by the middleware, or a zero value if the middleware did not run. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/pagination).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
