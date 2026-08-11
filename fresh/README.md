# fresh — implements HTTP response freshness testing, a port of the npm "fresh" package

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/fresh.svg)](https://pkg.go.dev/github.com/malcolmston/express/fresh)

Package fresh implements HTTP response freshness testing, a port of the npm
"fresh" package. It determines whether a cached response is still fresh
relative to a request's conditional headers, and is the same primitive that
Express uses to answer 304 Not Modified requests.

Use this package when a server holds a representation of a resource together
with its validators (an ETag and/or a Last-Modified time) and needs to decide
whether the client's cached copy is still current. When Fresh reports true,
the caller can skip sending the body and respond with 304 Not Modified; when
it reports false, the resource is considered stale and the full response
should be sent. This mirrors the conditional-request handling described in RFC
7232.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/fresh"
```

## Usage

This is the package's own `ExampleFresh`, so it compiles and its output is
asserted on every `go test ./fresh/`.

```go
reqHeaders := http.Header{}
	reqHeaders.Set("If-None-Match", `"abc123"`)

	resHeaders := http.Header{}
	resHeaders.Set("ETag", `"abc123"`)

	fmt.Println(fresh.Fresh(reqHeaders, resHeaders))
```

```
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Fresh(reqHeaders, resHeaders http.Header) bool` | Fresh reports whether a response is fresh given the request headers and the response headers. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/fresh).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
