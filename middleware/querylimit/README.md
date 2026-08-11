# querylimit — provides middleware that rejects requests whose raw query string exceeds a configured

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/querylimit.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/querylimit)

Package querylimit provides middleware that rejects requests whose raw query
string exceeds a configured maximum length. Oversized requests receive a 414
URI Too Long response before any route handler runs.

It plays the same defensive role as guards found around Node's query parsers
(for example the parameterLimit / maximum URL length caps in qs and
body-parser, and the URL-length limits enforced by reverse proxies): it bounds
the amount of query-string data the application will accept so that a single
request cannot force expensive parsing of an enormous or maliciously crafted
query. Rather than counting individual parameters, this port caps the total
byte length of the raw query string, which is a cheap, allocation-free check
performed before the query is parsed.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/querylimit"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./querylimit/`.

```go
app := express.New()
	app.Use(querylimit.New(querylimit.Options{MaxLength: 10}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	small := httptest.NewRecorder()
	app.ServeHTTP(small, httptest.NewRequest("GET", "/?a=1", nil))

	big := httptest.NewRecorder()
	app.ServeHTTP(big, httptest.NewRequest("GET", "/?q=aaaaaaaaaaaaaaaaaaaa", nil))

	fmt.Println(small.Code)
	fmt.Println(big.Code)
```

```
200
414
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns query-limit middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the query-limit middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/querylimit).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
