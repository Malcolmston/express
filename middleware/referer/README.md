# referer — provides middleware that captures the Referer request header, parses out its host

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/referer.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/referer)

Package referer provides middleware that captures the Referer request header,
parses out its host, and stores the result on the request under the key
"referer" for downstream handlers such as analytics and anti-CSRF checks. It
ports the small but recurring Node pattern of reading req.headers.referer (or
req.get('referrer')), normalizing it, and stashing it on the request/context
so later handlers need not re-parse the header themselves.

Use it when several downstream consumers care about where a request came from:
request logging and referral analytics, soft origin checks, or feeding a
stricter allowlist gate (see the sibling referercheck package). Doing the
parse once, up front, means handlers read a typed Referer value instead of
repeatedly pulling and url.Parse-ing a raw header, and it centralizes the
quirk that the header is misspelled "Referer" in the HTTP standard while some
clients and codebases use the correct "Referrer".

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/referer"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./referer/`.

```go
app := express.New()
	app.Use(referer.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		if ref, ok := referer.From(req); ok {
			fmt.Printf("url:  %s\n", ref.URL)
			fmt.Printf("host: %s\n", ref.Host)
		}
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Referer", "https://example.com/page?x=1")
	app.ServeHTTP(rec, req)
```

```
url:  https://example.com/page?x=1
host: example.com
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that stores a Referer via req.Set(Key, ref). |

### Types

| Type | What it is |
| --- | --- |
| `Referer` | Referer holds the raw Referer header and its parsed host. |

<details>
<summary><code>Referer</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func From(req *express.Request) (Referer, bool)` | From retrieves the Referer stored by this middleware. |

</details>

### Constants

`Key`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/referer).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
