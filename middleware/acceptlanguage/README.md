# acceptlanguage — provides middleware that parses the Accept-Language request header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/acceptlanguage.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/acceptlanguage)

Package acceptlanguage provides middleware that parses the Accept-Language
request header, optionally negotiates it against a set of supported languages,
and stores the chosen language tag on the request under the key "language". It
is the Go analogue of Node content-negotiation helpers such as the accepts /
negotiator packages and the accept-language module, exposed as a drop-in
express.Handler that resolves a single best language up front so handlers need
not reparse the header.

Use this middleware whenever an application serves localized content and needs
to know which language the client prefers. Rather than scattering header
parsing and locale-matching logic across handlers, mount this once and read
the resolved tag from the request. It is well suited to internationalized web
pages, localized API messages, and any route that selects templates, currency,
or copy by language, giving every downstream handler a consistent,
already-negotiated answer.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/acceptlanguage"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./acceptlanguage/`.

```go
app := express.New()
	app.Use(acceptlanguage.New(acceptlanguage.Options{
		Supported: []string{"en", "fr"},
		Default:   "en",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		lang, _ := req.Value(acceptlanguage.Key)
		res.Send(fmt.Sprintf("%v", lang))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "de,fr;q=0.9")
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
```

```
fr
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that stores the negotiated language via req.Set(Key, lang). |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the acceptlanguage middleware. |

### Constants

`Key`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/acceptlanguage).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
