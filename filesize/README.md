# filesize — converts a number of bytes into a human readable string

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/filesize.svg)](https://pkg.go.dev/github.com/malcolmston/express/filesize)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/filesize)

Package filesize converts a number of bytes into a human readable string, a
faithful port of the npm package "filesize".

Reach for this package whenever a raw byte count needs to be shown to a
person: file managers, upload progress, download sizes, disk usage, and log
output all read better as "1.34 kB" than as "1337". FileSize covers the common
case with sensible defaults, while FileSizeOpts exposes the knobs needed to
match a particular convention.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/filesize"
```

## Usage

This is the package's own `ExampleFileSizeOpts`, so it compiles and its output is
asserted on every `go test ./filesize/`.

```go
fmt.Println(filesize.FileSizeOpts(1337, filesize.Options{Base: 2}))
	fmt.Println(filesize.FileSizeOpts(1536, filesize.Options{Base: 2, Standard: "jedec"}))
```

```
1.31 KiB
1.5 KB
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func FileSize(bytes float64) string` | FileSize converts a number of bytes into a human readable string using the default options (base 10, two decimals). |
| `func FileSizeOpts(bytes float64, opts Options) string` | FileSizeOpts converts a number of bytes into a human readable string honouring the supplied Options. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures how a byte count is rendered. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/filesize).

## Measured parity

Compared case-for-case against **`filesize@11.0.22`** by the harness in
[`parity/express/nested/filesize`](https://github.com/malcolmston/go/tree/main/parity/express/nested/filesize):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 188 |
| Matching | 188 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/filesize/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
