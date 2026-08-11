# xid — generates globally unique, sortable 12-byte identifiers encoded as 20-character base32

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/xid.svg)](https://pkg.go.dev/github.com/malcolmston/express/xid)

Package xid generates globally unique, sortable 12-byte identifiers encoded as
20-character base32 strings. It is a standard-library-only Go port of the xid
scheme popularised by rs/xid and mirrored by the npm "xid-js" library. The
design follows MongoDB's ObjectID: it is compact, lexically sortable by
creation time, and safe to generate on many hosts and processes concurrently
without any central coordinator.

Each id is 12 raw bytes laid out as four fields. The first 4 bytes are a
big-endian Unix timestamp in seconds; the next 3 bytes are a machine
identifier; the following 2 bytes are the process id; and the final 3 bytes
are a monotonically increasing counter. Placing the timestamp in the most
significant bytes is what makes ids sort chronologically. The machine id is
derived from the first three bytes of the MD5 hash of the host name (falling
back to random bytes when the host name is unavailable), and both the machine
id and pid are captured once in the package init function.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/xid"
```

## Usage

This is the package's own `ExampleNewWithData`, so it compiles and its output is
asserted on every `go test ./xid/`.

```go
a := xid.NewWithData(1600000000, [3]byte{1, 2, 3}, 100, 5)
	b := xid.NewWithData(1600000000, [3]byte{1, 2, 3}, 100, 5)
	fmt.Println(a == b, len(a))
```

```
true 20
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Decode(id string) ([12]byte, error)` | Decode parses a 20-char xid string back into its 12 raw bytes. |
| `func New(unixSeconds int64) string` | New generates an xid for the given unix seconds using the process machine id, pid, and an atomically incrementing counter. |
| `func NewWithData(unixSeconds int64, machine [3]byte, p uint16, c uint32) string` | NewWithData builds an xid deterministically from all its components. |
| `func Time(id string) (int64, error)` | Time decodes the unix-seconds timestamp encoded in an xid string. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/xid).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
