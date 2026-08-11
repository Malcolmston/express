# snowflake — standard-library-only implementation of the Twitter Snowflake distributed ID scheme

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/snowflake.svg)](https://pkg.go.dev/github.com/malcolmston/express/snowflake)

Package snowflake is a standard-library-only implementation of the Twitter
Snowflake distributed ID scheme, echoing the design of popular npm libraries
such as "snowflake-id" and the Go package bwmarrin/snowflake. It generates
63-bit integer identifiers that are unique across many nodes, roughly ordered
by creation time, and cheap to produce at high throughput without any shared
state between machines beyond a distinct node id per generator.

Each id is a 63-bit value (it fits in a signed int64 and stays non-negative)
partitioned into three fields. The high 41 bits hold a millisecond timestamp
measured relative to a custom epoch; the middle 10 bits hold the node id
(0..1023); and the low 12 bits hold a per-millisecond sequence (0..4095). The
custom Epoch constant is the Twitter epoch, 1288834974657 ms (2010-11-04
01:42:54.657 UTC); subtracting it from the wall clock lets the 41-bit
timestamp field span about 69 years from that starting point rather than from
1970.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/snowflake"
```

## Usage

This is the package's own `ExampleNode_GenerateAt`, so it compiles and its output is
asserted on every `go test ./snowflake/`.

```go
node, _ := snowflake.NewNode(1)
	id := node.GenerateAt(snowflake.Epoch + 1000)
	fmt.Println(snowflake.NodeOf(id))
	fmt.Println(snowflake.Sequence(id))
	fmt.Println(snowflake.Timestamp(id) == snowflake.Epoch+1000)
```

```
1
0
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func NodeOf(id int64) int64` | NodeOf returns the node id encoded in id. |
| `func Sequence(id int64) int64` | Sequence returns the per-millisecond sequence encoded in id. |
| `func Timestamp(id int64) int64` | Timestamp returns the absolute Unix-millisecond timestamp encoded in id. |

### Types

| Type | What it is |
| --- | --- |
| `Node` | Node generates snowflake ids for a single node id. |

<details>
<summary><code>Node</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func NewNode(nodeID int64) (*Node, error)` | NewNode returns a Node for the given node id (0..1023). |
| `func (n *Node) Generate() int64` | Generate produces a new id using the current wall-clock time. |
| `func (n *Node) GenerateAt(ms int64) int64` | GenerateAt produces an id for the given millisecond timestamp. |

</details>

### Constants

`Epoch`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/snowflake).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
