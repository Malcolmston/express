# referrerpolicy — provides middleware that sets the Referrer-Policy response header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/referrerpolicy.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/referrerpolicy)

Package referrerpolicy provides middleware that sets the Referrer-Policy
response header, controlling how much referrer information the browser
attaches to outgoing requests originated from your pages. It is the express
port of Helmet's referrerPolicy middleware (helmet.referrerPolicy()), which in
turn wraps the single standardized header that browsers honor to govern the
Referer request header on navigations, resource loads, and prefetches.

Use it whenever you want to limit referrer leakage: the default Referer header
can disclose the full URL of the page a user came from, including path and
query string, to third-party origins and to downgraded (HTTPS to HTTP)
destinations. Setting an explicit policy such as "no-referrer" or
"strict-origin-when-cross-origin" is a common privacy and security baseline,
which is why it appears in Helmet and most hardened header presets. It is
cheap enough to enable globally with app.Use.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/referrerpolicy"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./referrerpolicy/`.

```go
app := express.New()
	app.Use(referrerpolicy.New(referrerpolicy.Options{
		Policy: []string{"no-referrer", "strict-origin-when-cross-origin"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Referrer-Policy"))
```

```
no-referrer, strict-origin-when-cross-origin
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the Referrer-Policy header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the referrerpolicy middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/referrerpolicy).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
