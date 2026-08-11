# useragent — provides middleware that performs lightweight parsing of the User-Agent request header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/useragent.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/useragent)

Package useragent provides middleware that performs lightweight parsing of the
User-Agent request header into a small struct and stores it on the request for
downstream handlers. It is a spiritual, dependency-free port of the Node
"express-useragent" middleware and the "useragent" npm package, but
deliberately tiny: rather than shipping a large table of regular expressions
covering hundreds of browsers and devices, it recognises only a handful of
common browser families and operating systems using plain substring matching.
The goal is to give downstream handlers a fast, coarse signal ("this looks
like mobile Chrome on Android") without pulling in any third-party parsing
library.

Reach for this middleware when you want to branch on broad client categories
— for example, serving a lighter template to mobile clients, recording
coarse browser statistics, or logging an OS breakdown — and can tolerate
approximate results. It is not suitable as a security control or for precise
device fingerprinting; the User-Agent header is client-supplied and easily
spoofed, and the detection here is intentionally shallow. For blocking
specific agents outright see the sibling useragentblock package, which this
package does not itself perform.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/useragent"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./useragent/`.

```go
app := express.New()
	app.Use(useragent.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		ua, _ := useragent.From(req)
		res.Send(fmt.Sprintf("browser=%s os=%s mobile=%t", ua.Browser, ua.OS, ua.Mobile))
	})

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 13; Pixel) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Mobile Safari/537.36")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Println(w.Body.String())
```

```
browser=Chrome os=Android mobile=true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that parses the User-Agent header and stores the resulting UserAgent via req.Set(Key, ua). |

### Types

| Type | What it is |
| --- | --- |
| `UserAgent` | UserAgent is the result of basic substring-based User-Agent detection. |

<details>
<summary><code>UserAgent</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func From(req *express.Request) (UserAgent, bool)` | From retrieves the parsed UserAgent stored by this middleware. |
| `func Parse(raw string) UserAgent` | Parse performs basic substring detection on a User-Agent string. |

</details>

### Constants

`Key`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/useragent).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
