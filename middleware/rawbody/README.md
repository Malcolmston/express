# rawbody — provides express middleware that reads the entire request body into memory

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/rawbody.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/rawbody)

Package rawbody provides express middleware that reads the entire request body
into memory, stores it on the request via SetBody, and restores req.Raw.Body
so downstream handlers can read it again.

It ports the buffering behaviour of Node's raw-body module (the same building
block body-parser uses under the hood): it drains the incoming request stream
once, into a single in-memory []byte, and makes that buffer available to the
rest of the application. Capturing the raw bytes is essential for tasks that
must see the exact payload — HMAC/webhook signature verification, request
logging or auditing, or custom parsing — where re-reading the network stream
would otherwise fail because an io.Reader can only be consumed once.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/rawbody"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./rawbody/`.

```go
app := express.New()
	app.Use(rawbody.New())
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		captured, _ := req.Body().([]byte)
		reread, _ := io.ReadAll(req.Raw.Body)
		fmt.Printf("captured=%q reread=%q\n", captured, reread)
		res.Send("ok")
	})

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("POST", "/", strings.NewReader("hello world")))
```

```
captured="hello world" reread="hello world"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that buffers the full request body into a []byte and makes it available through req.Body(). |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/rawbody).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
