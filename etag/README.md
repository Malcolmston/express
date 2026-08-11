# etag — creates HTTP entity tags, a port of the npm "etag" package

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/etag.svg)](https://pkg.go.dev/github.com/malcolmston/express/etag)

Package etag creates HTTP entity tags, a port of the npm "etag" package.

An entity tag (ETag) is an opaque validator that a server sends in the ETag
response header so that clients can make conditional requests with
If-None-Match or If-Match. When the representation of a resource changes its
ETag changes, allowing caches and browsers to revalidate cheaply and servers
to answer unchanged requests with 304 Not Modified instead of resending the
body.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/etag"
```

## Usage

This is the package's own `ExampleGenerateStat`, so it compiles and its output is
asserted on every `go test ./etag/`.

```go
modtime := time.Unix(1, 0).UTC()
	tag := etag.GenerateStat(1024, modtime, false)
	fmt.Println(tag)
```

```
"400-3e8"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Generate(data []byte, weak bool) string` | Generate computes an entity tag for the given content. |
| `func GenerateStat(size int64, modtime time.Time, weak bool) string` | GenerateStat computes a stat-based entity tag from a resource's size and modification time. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/etag).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
