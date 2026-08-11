# serveindex — provides middleware that renders an HTML directory listing for requests that resolve

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/serveindex.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/serveindex)

Package serveindex provides middleware that renders an HTML directory listing
for requests that resolve to a directory beneath a configured root. It is the
express framework's Go analogue of the Node connect/express serve-index
middleware, which produces the familiar "Index of /path" pages, reduced here
to a compact, dependency-free HTML renderer built on the standard library's
os.ReadDir and html.EscapeString.

Reach for this middleware to expose a browsable view of a directory tree — a
downloads area, a build-artifact folder, static documentation, or a quick file
share — where users benefit from clickable links to sub-directories and
files. It is designed to sit next to a static file server: serveindex renders
the folder pages while the file server delivers the actual file contents, so
the two together reproduce the classic auto-index behaviour of a plain web
server.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/serveindex"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./serveindex/`.

```go
dir, _ := os.MkdirTemp("", "serveindex-example")
	defer os.RemoveAll(dir)
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644)
	_ = os.Mkdir(filepath.Join(dir, "sub"), 0o755)

	app := express.New()
	app.Use(serveindex.New(serveindex.Options{Root: dir}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	body := w.Body.String()
	fmt.Println(strings.HasPrefix(w.Header().Get("Content-Type"), "text/html"))
	fmt.Println(strings.Contains(body, `<a href="/sub/">sub/</a>`))
	fmt.Println(strings.Contains(body, `<a href="/a.txt">a.txt</a>`))
```

```
true
true
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that serves directory listings from Options.Root. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the serveindex middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/serveindex).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
