# useragentblock — provides middleware that rejects requests whose User-Agent header contains any

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/useragentblock.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/useragentblock)

Package useragentblock provides middleware that rejects requests whose
User-Agent header contains any configured blocked substring. It is a small,
dependency-free equivalent of the Node "express-useragent-blocker" family of
middleware and the hand-rolled "block by user-agent" guards commonly written
on top of connect: a deny-list applied to the client-supplied User-Agent
header. There is no upstream npm package this maps to one-for-one; it captures
the widespread pattern of denying obvious scraper, bot and CLI tooling agents
by name.

Use it to shed unwanted traffic cheaply — blocking greedy crawlers, curl or
wget probes, headless scrapers, or a specific misbehaving bot — before that
traffic reaches expensive downstream handlers. Because the User-Agent header
is trivially spoofed, treat this as traffic hygiene and coarse abuse
mitigation rather than a real security boundary; a determined client can
always send a different header. For richer, non-blocking classification of the
same header see the sibling useragent package.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/useragentblock"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./useragentblock/`.

```go
app := express.New()
	app.Use(useragentblock.New(useragentblock.Options{
		Block: []string{"curl", "BadBot"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	do := func(ua string) int {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("User-Agent", ua)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	fmt.Println("blocked:", do("CURL/7.0"))
	fmt.Println("allowed:", do("Mozilla/5.0 Firefox/120.0"))
```

```
blocked: 403
allowed: 200
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that responds with 403 when the request's User-Agent contains one of the configured blocked substrings. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the user-agent blocking middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/useragentblock).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
