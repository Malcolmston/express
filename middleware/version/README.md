# version — provides middleware that exposes an application's version

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/version.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/version)

Package version provides middleware that exposes an application's version. It
sets an X-Version header on every response and, when a request targets the
configured path, replies with a small JSON document reporting the version.
There is no single canonical Node package this ports; it distils a common
operational pattern seen across Express deployments — advertising the
running build both as a response header and via a dedicated introspection
endpoint (comparable to hand-rolled "/version" routes and helpers like
"express-version-route") — into one small, dependency-free middleware.

Use it to make deploys observable: a load balancer, uptime probe, or curl
against the version path can confirm which build is live, and every ordinary
response carries the version header so clients and logs can correlate
behaviour with a release. It pairs naturally with a healthcheck endpoint.
Because the version string is emitted to all callers, treat it as public
information and avoid encoding anything sensitive in it.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/version"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./version/`.

```go
app := express.New()
	app.Use(version.New(version.Options{Version: "1.2.3"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("home")
	})

	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, httptest.NewRequest("GET", "/version", nil))
	fmt.Println("body:", w1.Body.String())

	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, httptest.NewRequest("GET", "/", nil))
	fmt.Println("home:", w2.Body.String())
	fmt.Println("header:", w2.Header().Get("X-Version"))
```

```
body: {"version":"1.2.3"}
home: home
header: 1.2.3
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns version middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the version middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/version).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
