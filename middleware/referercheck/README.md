# referercheck — provides middleware that rejects requests whose Referer header host is not

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/referercheck.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/referercheck)

Package referercheck provides middleware that rejects requests whose Referer
header host is not in a configured allowlist. It ports the classic Node
referer-based CSRF guard — the pattern behind packages and hand-written
middleware that compare req.headers.referer against a set of permitted hosts
and 403 anything else — into a single express handler backed by a host set.

Use it as a lightweight, defense-in-depth guard on state-changing endpoints
(form posts, mutating APIs) to ensure the request was initiated from a page
your own site served. Because a browser sets the Referer for same-site
navigations, requiring a known host raises the bar for cross-site request
forgery and hotlinking without the token plumbing of a full CSRF library. It
is best treated as one layer among several rather than a sole defense, since
the Referer header can be suppressed or stripped by proxies and privacy tools.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/referercheck"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./referercheck/`.

```go
app := express.New()
	app.Use(referercheck.New(referercheck.Options{
		Allow: []string{"example.com"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	probe := func(ref string) int {
		r := httptest.NewRequest("GET", "/", nil)
		if ref != "" {
			r.Header.Set("Referer", ref)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	fmt.Println("allowed:   ", probe("https://example.com/page"))
	fmt.Println("disallowed:", probe("https://evil.com/page"))
	fmt.Println("missing:   ", probe(""))
```

```
allowed:    200
disallowed: 403
missing:    403
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that responds with 403 unless the request's Referer header host appears in the allowlist. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the referer-check middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/referercheck).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
