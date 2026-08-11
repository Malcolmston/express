# originagentcluster — provides middleware that sets the Origin-Agent-Cluster response header on every

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/originagentcluster.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/originagentcluster)

Package originagentcluster provides middleware that sets the
Origin-Agent-Cluster response header on every response, requesting that the
browser place the origin into its own, origin-keyed agent cluster. It is a
direct port of Helmet's originAgentCluster middleware, which emits the same
single, fixed header value and takes no configuration.

The Origin-Agent-Cluster header is a hint to the user agent that documents
from this origin should be isolated from other origins that happen to share
the same site (scheme plus registrable domain). When honored, the browser keys
the agent cluster by origin rather than by site, which prevents same-site
pages from reaching into this origin's synchronous scripting context via
mechanisms such as document.domain and can give the origin its own dedicated
process. Use it as a defense-in-depth hardening measure on applications that
never rely on cross-origin, same-site DOM access.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/originagentcluster"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./originagentcluster/`.

```go
app := express.New()
	app.Use(originagentcluster.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("hello")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Header().Get("Origin-Agent-Cluster"))
```

```
?1
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that sets Origin-Agent-Cluster: ?1. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/originagentcluster).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
