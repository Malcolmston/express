# nanoid — generates compact, URL-safe, cryptographically random identifiers

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/nanoid.svg)](https://pkg.go.dev/github.com/malcolmston/express/nanoid)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/nanoid)

Package nanoid generates compact, URL-safe, cryptographically random
identifiers, a Go port of the npm "nanoid" package. nanoid is a popular modern
alternative to UUIDs: the identifiers are shorter, need no hyphenation or
special encoding to travel safely in URLs, and are produced from a
cryptographically secure random source. This implementation uses only the Go
standard library, drawing its randomness from crypto/rand.

You use this package whenever you need unique, unguessable, opaque tokens —
database primary keys, short links, session or resource identifiers, file
names — and want something more compact and URL-friendly than a canonical
UUID string. Because the bytes come from a CSPRNG the ids are suitable for
contexts where predictability would be a security problem, and because the
default alphabet is URL- and filename-safe they can be dropped into a path or
query string without escaping.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/nanoid"
```

## Usage

This is the package's own `ExampleCustom`, so it compiles and its output is
asserted on every `go test ./nanoid/`.

```go
id, err := nanoid.Custom("0123456789abcdef", 16)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(id))
	fmt.Println(strings.Trim(id, "0123456789abcdef") == "")
```

```
16
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Custom(alphabet string, size int) (string, error)` | Custom returns a nanoid using the given alphabet and size, sampling bytes from crypto/rand with an unbiased mask/reject method. |
| `func New() (string, error)` | New returns a nanoid using the default alphabet and size. |
| `func NewSize(size int) (string, error)` | NewSize returns a nanoid using the default alphabet and the given size. |

### Constants

`DefaultAlphabet`, `DefaultSize`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/nanoid).

## Measured parity

Compared case-for-case against **`nanoid@6.0.1`** by the harness in
[`parity/express/nested/nanoid`](https://github.com/malcolmston/go/tree/main/parity/express/nested/nanoid):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 126 |
| Matching | 123 |
| Mismatching | 0 |
| Declared deviations | 3 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/nanoid/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
