# crossoriginopener — provides express middleware that sets the Cross-Origin-Opener-Policy (COOP) response

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/crossoriginopener.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/crossoriginopener)

Package crossoriginopener provides express middleware that sets the
Cross-Origin-Opener-Policy (COOP) response header on every response. It is the
Go analogue of helmet's crossOriginOpenerPolicy feature (from the Node helmet
package), packaged as a drop-in express.Handler. The one header it writes, in
full, is "Cross-Origin-Opener-Policy".

COOP controls whether a document may share a browsing context group — and
therefore a JavaScript reference via window.opener — with cross-origin
documents it opens or is opened by. Setting it to same-origin severs that
relationship for cross-origin windows, which both hardens the page against
cross-window attacks such as tabnabbing and XS-Leaks and, together with a
suitable Cross-Origin-Embedder-Policy, is one of the two headers required to
make a document cross-origin isolated (unlocking SharedArrayBuffer and precise
timers). Reach for this middleware to isolate your pages from the windows they
interact with, or as part of enabling cross-origin isolation.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/crossoriginopener"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./crossoriginopener/`.

```go
app := express.New()
	app.Use(crossoriginopener.New(crossoriginopener.Options{Policy: "same-origin-allow-popups"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Cross-Origin-Opener-Policy"))
```

```
same-origin-allow-popups
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the Cross-Origin-Opener-Policy header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the crossoriginopener middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/crossoriginopener).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
