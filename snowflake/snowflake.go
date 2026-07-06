// Package snowflake is a standard-library-only implementation of the Twitter
// Snowflake distributed ID scheme, echoing the design of popular npm libraries
// such as "snowflake-id" and the Go package bwmarrin/snowflake. It generates
// 63-bit integer identifiers that are unique across many nodes, roughly ordered
// by creation time, and cheap to produce at high throughput without any shared
// state between machines beyond a distinct node id per generator.
//
// Each id is a 63-bit value (it fits in a signed int64 and stays non-negative)
// partitioned into three fields. The high 41 bits hold a millisecond timestamp
// measured relative to a custom epoch; the middle 10 bits hold the node id
// (0..1023); and the low 12 bits hold a per-millisecond sequence (0..4095). The
// custom Epoch constant is the Twitter epoch, 1288834974657 ms
// (2010-11-04 01:42:54.657 UTC); subtracting it from the wall clock lets the
// 41-bit timestamp field span about 69 years from that starting point rather
// than from 1970.
//
// A Node bundles a node id with the small amount of mutable state needed to
// assign sequence numbers, guarded by a mutex so a single Node is safe for
// concurrent use. NewNode validates the node id against the 10-bit range and
// returns an error if it is out of bounds. GenerateAt composes an id for an
// explicit millisecond timestamp, and Generate is the convenience wrapper that
// calls GenerateAt with time.Now().UnixMilli().
//
// Monotonicity is handled by the sequence field. When GenerateAt is called
// repeatedly within the same millisecond, the sequence increments so each id is
// distinct and strictly increasing; if the 12-bit sequence overflows (more than
// 4096 ids in one millisecond), the generator advances its internal timestamp
// by one millisecond and continues, effectively borrowing from the future to
// preserve ordering. When the clock moves forward to a new millisecond the
// sequence resets to zero. Because the timestamp occupies the most significant
// bits, ids from a single node increase over time and are broadly sortable
// across nodes, though ids minted in the same millisecond on different nodes are
// ordered by node id rather than by true creation instant.
//
// The package-level helpers Timestamp, NodeOf, and Sequence decompose an id
// back into its absolute Unix-millisecond time, node id, and sequence. Compared
// with the reference JavaScript and Go implementations, the 41/10/12 bit layout,
// the Twitter epoch, and the overflow-then-advance behaviour match; the main
// differences are the idiomatic Go Node type with explicit GenerateAt for
// deterministic testing and the plain int64 return type instead of a string or
// BigInt wrapper.
package snowflake

import (
	"errors"
	"sync"
	"time"
)

const (
	// Epoch is the custom epoch in Unix milliseconds (the Twitter epoch,
	// 2010-11-04 01:42:54.657 UTC).
	Epoch int64 = 1288834974657

	nodeBits uint8 = 10
	stepBits uint8 = 12

	nodeMax int64 = (1 << nodeBits) - 1
	stepMax int64 = (1 << stepBits) - 1

	timeShift uint8 = nodeBits + stepBits
	nodeShift uint8 = stepBits
)

// Node generates snowflake ids for a single node id.
type Node struct {
	mu   sync.Mutex
	id   int64
	time int64
	step int64
}

// NewNode returns a Node for the given node id (0..1023).
func NewNode(nodeID int64) (*Node, error) {
	if nodeID < 0 || nodeID > nodeMax {
		return nil, errors.New("snowflake: node id must be between 0 and 1023")
	}
	return &Node{id: nodeID}, nil
}

// GenerateAt produces an id for the given millisecond timestamp. Repeated calls
// with the same timestamp increment the sequence; on sequence overflow the
// timestamp is advanced by one millisecond.
func (n *Node) GenerateAt(ms int64) int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	if ms == n.time {
		n.step = (n.step + 1) & stepMax
		if n.step == 0 {
			ms = n.time + 1
			n.time = ms
		}
	} else {
		n.step = 0
		n.time = ms
	}

	return ((ms - Epoch) << timeShift) | (n.id << nodeShift) | n.step
}

// Generate produces a new id using the current wall-clock time.
func (n *Node) Generate() int64 {
	return n.GenerateAt(time.Now().UnixMilli())
}

// Timestamp returns the absolute Unix-millisecond timestamp encoded in id.
func Timestamp(id int64) int64 {
	return (id >> timeShift) + Epoch
}

// NodeOf returns the node id encoded in id.
func NodeOf(id int64) int64 {
	return (id >> nodeShift) & nodeMax
}

// Sequence returns the per-millisecond sequence encoded in id.
func Sequence(id int64) int64 {
	return id & stepMax
}
