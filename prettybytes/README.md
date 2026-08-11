# prettybytes — converts a number of bytes into a compact

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/prettybytes.svg)](https://pkg.go.dev/github.com/malcolmston/express/prettybytes)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/prettybytes)

Package prettybytes converts a number of bytes into a compact, human readable
string such as "1.34 kB" or "1.5 MiB". It is a faithful port of the npm
package "pretty-bytes", reproducing its unit tables, rounding and formatting
rules using only the Go standard library (math, strconv and strings).

Raw byte counts are hard for people to read: "1337000000" conveys far less at
a glance than "1.34 GB". This package is what you use to render file sizes,
transfer amounts, memory footprints and similar quantities in logs, CLIs and
UIs. The default entry point PrettyBytes takes a float64 count and returns the
SI (base-1000) rendering; PrettyBytesOpts takes the same number plus an
Options struct for finer control.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/prettybytes"
```

## Usage

This is the package's own `ExamplePrettyBytesOpts`, so it compiles and its output is
asserted on every `go test ./prettybytes/`.

```go
fmt.Println(prettybytes.PrettyBytesOpts(1024, prettybytes.Options{Binary: true}))
	fmt.Println(prettybytes.PrettyBytesOpts(1337, prettybytes.Options{Bits: true}))
```

```
1 KiB
1.34 kbit
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func PrettyBytes(n float64) string` | PrettyBytes converts a number of bytes into a human readable string using SI units (base 1000). |
| `func PrettyBytesOpts(number float64, opts Options) string` | PrettyBytesOpts converts a number of bytes into a human readable string honouring the supplied Options. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures how a byte count is rendered. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/prettybytes).

## Measured parity

Compared case-for-case against **`pretty-bytes@7.1.1`** by the harness in
[`parity/express/nested/prettybytes`](https://github.com/malcolmston/go/tree/main/parity/express/nested/prettybytes):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 190 |
| Matching | 187 |
| Mismatching | 0 |
| Declared deviations | 3 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/prettybytes/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
