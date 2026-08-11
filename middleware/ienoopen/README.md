# ienoopen — provides middleware that sets the X-Download-Options response header to the single

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/ienoopen.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/ienoopen)

Package ienoopen provides middleware that sets the X-Download-Options response
header to the single value "noopen". It is a direct port of the Helmet
"ienoopen" / X-Download-Options component (the same behavior Helmet exposes
today through helmet.ieNoOpen()), reproducing that header in a stdlib-only
Express-style stack for Go.

The header exists to defend a very specific legacy behavior in Internet
Explorer 8 and later. Historically those browsers offered an "Open" button in
the file-download dialog that executed a downloaded file directly in the
security context of the originating site. For an application that serves
user-supplied or otherwise untrusted downloads, that behavior allowed an
attacker-controlled file to run as if it came from your origin. Sending
X-Download-Options: noopen removes the "Open" option, forcing users to save
the file to disk before opening it. Use this middleware when your service
hands out downloadable content and you still care about hardening older IE
clients; on modern browsers the header is simply ignored and harmless.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/ienoopen"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./ienoopen/`.

```go
app := express.New()
	app.Use(ienoopen.New())
	app.Get("/download", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("file-contents")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/download", nil))

	fmt.Println(rec.Header().Get("X-Download-Options"))
```

```
noopen
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that sets X-Download-Options: noopen. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/ienoopen).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
