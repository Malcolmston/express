# uuid — generates and parses RFC 4122 universally unique identifiers

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/uuid.svg)](https://pkg.go.dev/github.com/malcolmston/express/uuid)

Package uuid generates and parses RFC 4122 universally unique identifiers. It
is a small, standard-library-only Go port of the npm "uuid" package, exposing
the two identifier flavours that the vast majority of applications need: fully
random version 4 UUIDs and name-based version 5 (SHA-1) UUIDs. A UUID is a
128-bit value canonically rendered as 36 lowercase hexadecimal characters in
the grouping 8-4-4-4-12 (for example "6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
and every value produced here sets the RFC 4122 variant bits (10xx) and the
appropriate version nibble.

V4 is the workhorse: it reads 16 bytes from crypto/rand, forces the version
nibble to 4 and the variant bits to the RFC 4122 layout, then formats the
result. Because 122 of the 128 bits are random, collisions are astronomically
unlikely and no coordination between generators is required. Use V4 whenever
you need an opaque, unguessable identifier and have no reason to derive it
from an existing name.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/uuid"
```

## Usage

This is the package's own `ExampleParse`, so it compiles and its output is
asserted on every `go test ./uuid/`.

```go
b, err := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		panic(err)
	}
	fmt.Println(uuid.Format(b))
```

```
6ba7b810-9dad-11d1-80b4-00c04fd430c8
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Format(b [16]byte) string` | Format renders a 16-byte array as a canonical lowercase UUID string. |
| `func Parse(s string) ([16]byte, error)` | Parse parses a UUID string into a 16-byte array. |
| `func V4() (string, error)` | V4 returns a randomly generated UUID version 4 string. |
| `func V5(namespace string, name string) (string, error)` | V5 returns a name-based UUID version 5 (SHA-1) string. |
| `func Validate(s string) bool` | Validate reports whether s is a well-formed UUID string. |

### Constants

`NamespaceDNS`, `NamespaceURL`, `NamespaceOID`, `NamespaceX500`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/uuid).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
