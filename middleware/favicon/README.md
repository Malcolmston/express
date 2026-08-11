# favicon — provides express middleware that answers requests for /favicon.ico

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/favicon.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/favicon)

Package favicon provides express middleware that answers requests for
/favicon.ico, short-circuiting them before they reach application routes or
logging. It ports the Node "serve-favicon" middleware in spirit: browsers
request /favicon.ico automatically and often, and this handler deals with
those requests in one place instead of letting them fall through to routers,
loggers, and not-found handlers. Unlike serve-favicon it can also operate
without any icon at all, replying 204 No Content.

Use this middleware to keep favicon noise out of application logic and access
logs, to cheaply serve a small embedded icon, or to explicitly answer favicon
probes with 204 when you have no icon to serve. Because it short-circuits,
mount it early — typically before your logging middleware and route handlers
— so favicon requests are handled and terminated before any of that runs.
The request path it claims is exported as the Path constant, "/favicon.ico".

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/favicon"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./favicon/`.

```go
app := express.New()
	app.Use(favicon.New(favicon.Options{
		Data:   []byte{0x00, 0x00, 0x01, 0x00},
		MaxAge: 3600,
	}))
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("app")
	})

	icon := httptest.NewRecorder()
	app.ServeHTTP(icon, httptest.NewRequest(http.MethodGet, favicon.Path, nil))
	fmt.Println(icon.Code, icon.Header().Get("Content-Type"))

	page := httptest.NewRecorder()
	app.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/index.html", nil))
	fmt.Println(page.Body.String())
```

```
200 image/x-icon
app
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that handles GET/HEAD requests for /favicon.ico and passes every other request through to next. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the favicon middleware. |

### Constants

`Path`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/favicon).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
