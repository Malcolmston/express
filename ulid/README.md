# ulid — generates and decodes Universally Unique Lexicographically Sortable Identifiers (ULIDs)

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/ulid.svg)](https://pkg.go.dev/github.com/malcolmston/express/ulid)

Package ulid generates and decodes Universally Unique Lexicographically
Sortable Identifiers (ULIDs). It is a standard-library-only Go port of the npm
"ulid" package. A ULID packs a millisecond-precision timestamp together with
random entropy into a single 128-bit value that is both globally unique and
naturally ordered by creation time, giving you the k-sortability of an
auto-increment key without a central coordinator.

The 128 bits are split into two fields: the high 48 bits hold a Unix timestamp
in milliseconds, and the low 80 bits hold random entropy. New reads those 10
entropy bytes from crypto/rand, while NewWithEntropy lets the caller supply
them for deterministic or reproducible generation. The timestamp field is 48
bits wide, so New/NewWithEntropy reject any ms value greater than 2^48-1
(roughly the year 10889), and the entropy slice must be exactly ten bytes
long.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/ulid"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./ulid/`.

```go
id, err := ulid.New(1234567890123)
	if err != nil {
		panic(err)
	}
	ts, _ := ulid.Timestamp(id)
	fmt.Println(len(id), ts == 1234567890123)
```

```
26 true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Decode(id string) ([16]byte, error)` | Decode parses a 26-char ULID string into its 16 raw bytes. |
| `func New(ms uint64) (string, error)` | New generates a ULID for the given millisecond timestamp with random entropy. |
| `func NewWithEntropy(ms uint64, entropy []byte) (string, error)` | NewWithEntropy generates a ULID deterministically from ms and 10-byte entropy. |
| `func Timestamp(id string) (uint64, error)` | Timestamp decodes the millisecond timestamp encoded in a ULID string. |

### Constants

`Alphabet`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/ulid).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
