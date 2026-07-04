// Package snowflake is a standard-library implementation of the Twitter
// Snowflake distributed ID scheme: a 63-bit id composed of a 41-bit
// millisecond timestamp (relative to a custom epoch), a 10-bit node id and a
// 12-bit per-millisecond sequence.
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
