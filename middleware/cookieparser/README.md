# cookieparser — provides express middleware that parses every cookie from the incoming request into

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/cookieparser.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/cookieparser)

Package cookieparser provides express middleware that parses every cookie from
the incoming request into a map`string`string and stores it on the request for
convenient downstream access. It is the Go analogue of the Node cookie-parser
middleware (expressjs/cookie-parser), packaged as a drop-in express.Handler,
and it turns the raw Cookie request header into a lookup map so handlers need
not re-parse it themselves.

Use this middleware whenever downstream handlers or other middleware want to
read cookies by name — session identifiers, feature flags, CSRF tokens,
preferences, and the like. Mount it once near the top of the chain with
app.Use so the parsed map is available everywhere, or attach it to a specific
router or path prefix if only part of the tree needs cookie access. It only
reads request cookies; setting cookies on the response remains the job of
res.Cookie.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/cookieparser"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./cookieparser/`.

```go
app := express.New()
	app.Use(cookieparser.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		cookies := cookieparser.From(req)
		fmt.Println(cookies["session"])
		fmt.Println(cookies["greeting"])
		res.Send("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	req.AddCookie(&http.Cookie{Name: "greeting", Value: "hello%20world"})
	app.ServeHTTP(httptest.NewRecorder(), req)
```

```
abc123
hello world
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func From(req *express.Request) map[string]string` | From returns the parsed cookie map previously stored by the middleware. |
| `func New() express.Handler` | New returns middleware that parses every cookie on the request into a map`string`string and stores it via req.Set("cookies", m). |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/cookieparser).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
