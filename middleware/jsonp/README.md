# jsonp — provides Express middleware that wraps JSON responses in a JavaScript callback

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/jsonp.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/jsonp)

Package jsonp provides Express middleware that wraps JSON responses in a
JavaScript callback invocation when the request carries a callback query
parameter, enabling classic cross-origin JSONP requests. It reproduces the
JSONP behavior built into Express's own res.jsonp() (which reads the callback
name from the "callback" query parameter by default) as a standalone,
stdlib-only handler that transparently upgrades any JSON body produced
downstream.

JSONP ("JSON with padding") predates CORS: a page loads data from another
origin by including it as a <script src="...?callback=fn">, and the server
replies with executable JavaScript — fn({...}) — that calls a function the
page defined. Use this middleware when you must support legacy browsers or
legacy clients that still rely on that <script>-tag technique and cannot use
CORS or fetch. For modern applications CORS is the correct choice; JSONP is
inherently a script-injection channel and should be treated with care.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/jsonp"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./jsonp/`.

```go
app := express.New()
	app.Use(jsonp.New())
	app.Get("/data", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON(map[string]int{"a": 1})
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/data?callback=cb", nil))

	fmt.Println(rec.Body.String())
	fmt.Println(rec.Header().Get("Content-Type"))
```

```
cb({"a":1});
application/javascript; charset=utf-8
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that, when a valid callback query parameter is present, transforms the response body from `<json>` into `callback(<json>);`… |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the jsonp middleware. |

### Constants

`DefaultParam`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/jsonp).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
