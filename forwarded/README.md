# forwarded — parses the X-Forwarded-For header along with a socket remote address to produce

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/forwarded.svg)](https://pkg.go.dev/github.com/malcolmston/express/forwarded)

Package forwarded parses the X-Forwarded-For header along with a socket remote
address to produce the chain of addresses that a request traversed. It is a Go
port of the npm module "forwarded" (the same helper that Express and the
"proxy-addr" module build on) and reproduces that module's ordering and
trimming rules exactly.

Use this package when your application sits behind one or more reverse proxies
or load balancers and you need to recover the client's original address, or
the full list of intermediaries a request passed through. The socket's
immediate peer (the nearest proxy, or the client itself when there is no
proxy) is always known reliably; the X-Forwarded-For header records the
addresses each proxy observed and appended as the request was relayed onward.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/forwarded"
```

## Usage

This is the package's own `ExampleFromRequest`, so it compiles and its output is
asserted on every `go test ./forwarded/`.

```go
r, _ := http.NewRequest("GET", "http://example.com", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")

	fmt.Println(forwarded.FromRequest(r))
```

```
[127.0.0.1 10.0.0.2 10.0.0.1]
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Forwarded(remoteAddr string, xff string) []string` | Forwarded returns the list of addresses in the request chain. |
| `func FromRequest(r *http.Request) []string` | FromRequest returns the address chain for an *http.Request, reading r.RemoteAddr (with the port stripped) and the X-Forwarded-For header. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/forwarded).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
