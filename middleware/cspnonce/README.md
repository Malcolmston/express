# cspnonce — provides middleware that generates a fresh

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/cspnonce.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/cspnonce)

Package cspnonce provides middleware that generates a fresh, cryptographically
random nonce for every request and emits a Content-Security-Policy header
whose script-src directive includes that nonce, enabling specific inline
scripts to be allow-listed safely. It complements the sibling csp package:
where csp sets a static policy from a directive map, cspnonce specializes in
the nonce workflow that Helmet's contentSecurityPolicy exposes through
per-request directive functions in the Node ecosystem, producing a
"'nonce-<value>'" source on each response.

Use this middleware when a page must include inline <script> blocks but you
still want the protection of a strict Content-Security-Policy. A nonce lets
the browser execute exactly the inline scripts whose nonce attribute matches
the one advertised in the header, while blocking any injected script that
lacks it — the recommended modern alternative to "'unsafe-inline'". Because
a new nonce is minted per request and never reused, an attacker cannot predict
or replay it. Mount it with app.Use ahead of the handlers that render HTML, so
each of those handlers can read the current nonce and stamp it onto the tags
it emits.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/cspnonce"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./cspnonce/`.

```go
app := express.New()
	app.Use(cspnonce.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		nonce := cspnonce.Nonce(req)
		res.Send("<script nonce=\"" + nonce + "\">/* inline */</script>")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	csp := rec.Header().Get("Content-Security-Policy")
	fmt.Println(strings.HasPrefix(csp, "default-src 'self'; script-src 'self' 'nonce-"))
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that generates a nonce, exposes it via req.Set("nonce", n) and res.Locals["nonce"], and sets a Content-Security-Policy header… |
| `func Nonce(req *express.Request) string` | Nonce returns the nonce generated for the request, or "" if the middleware did not run. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the CSP nonce middleware. |

### Constants

`ContextKey`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/cspnonce).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
