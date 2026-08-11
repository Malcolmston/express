# deburr — provides a faithful port of lodash's deburr utility

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/deburr.svg)](https://pkg.go.dev/github.com/malcolmston/express/deburr)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/deburr)

Package deburr provides a faithful port of lodash's deburr utility, built on
only the Go standard library. It exposes the single Deburr function, which
converts accented Latin letters to their basic ASCII equivalents and removes
combining diacritical marks, so that text such as "déjà vu" becomes "deja
vu" and "Crème Brûlée" becomes "Creme Brulee".

Deburring is most useful as a normalization step ahead of slugging, searching,
sorting, or generating identifiers. Turning "Málaga" into "Malaga" lets a URL
slug stay ASCII, lets a case-insensitive search match regardless of the
accents a user typed, and lets names sort in a predictable order without
pulling in a full Unicode collation library. It is a lossy, presentation-level
transform, not a transliteration engine: it approximates the Latin letter a
reader would recognize rather than attempting a linguistically correct
romanization.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/deburr"
```

## Usage

This is the package's own `ExampleDeburr`, so it compiles and its output is
asserted on every `go test ./deburr/`.

```go
fmt.Println(deburr.Deburr("déjà vu — Crème Brûlée"))
```

```
deja vu — Creme Brulee
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Deburr(s string) string` | Deburr converts Latin-1 Supplement and Latin Extended-A letters in s to their basic Latin equivalents and removes combining diacritical marks. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/deburr).

## Measured parity

Compared case-for-case against **`lodash@4.17.21`** by the harness in
[`parity/express/nested/deburr`](https://github.com/malcolmston/go/tree/main/parity/express/nested/deburr):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 81 |
| Matching | 81 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/deburr/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
