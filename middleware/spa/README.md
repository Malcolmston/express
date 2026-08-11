# spa — provides middleware for serving single-page applications

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/spa.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/spa)

Package spa provides middleware for serving single-page applications. It
serves static files from a root directory and, for navigation requests that do
not resolve to a real file, falls back to the SPA's index document so that
client-side routing works on deep links and page reloads. It is the express
framework's Go analogue of the connect-history-api-fallback middleware
combined with a static file server (express.static / serve-static): the
classic "serve the build output, and rewrite unknown routes to index.html"
pattern used by React, Vue, Angular, and Svelte production bundles.

Reach for this middleware to host a compiled front end from the same Go
process that serves its API. Without it, a user who reloads the page at
/dashboard/settings or shares a deep link would hit the server with a path
that has no matching file and receive a 404, because the routing for that path
lives in JavaScript, not on disk. This middleware bridges that gap by
returning index.html for such navigation requests so the client router can
take over and render the right view.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/spa"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./spa/`.

```go
root, _ := os.MkdirTemp("", "spa-demo")
	defer os.RemoveAll(root)
	os.WriteFile(filepath.Join(root, "index.html"), []byte("APP-SHELL"), 0o644)
	os.WriteFile(filepath.Join(root, "app.js"), []byte("console.log(1)"), 0o644)

	app := express.New()
	app.Use(spa.New(spa.Options{Root: root}))
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(404).Send("not found")
	})

	get := func(path string) (int, string) {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w.Code, w.Body.String()
	}

	c1, b1 := get("/app.js")
	c2, b2 := get("/dashboard/settings")
	c3, b3 := get("/missing.css")
	fmt.Printf("asset:    %d %s\n", c1, b1)
	fmt.Printf("deeplink: %d %s\n", c2, b2)
	fmt.Printf("missing:  %d %s\n", c3, b3)
```

```
asset:    200 console.log(1)
deeplink: 200 APP-SHELL
missing:  404 not found
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that serves files from Root, falling back to Index. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the spa middleware. |

### Constants

`DefaultIndex`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/spa).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
