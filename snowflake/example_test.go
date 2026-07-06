package snowflake_test

import (
	"fmt"

	"github.com/malcolmston/express/snowflake"
)

// ExampleNode_GenerateAt builds an id for an explicit millisecond timestamp so
// the result is fully deterministic and can be decomposed. GenerateAt is used
// instead of Generate (which reads the wall clock) precisely so the example does
// not depend on the current time. The package-level helpers then extract the
// three fields packed into the 63-bit value: the node id, the per-millisecond
// sequence, and the absolute timestamp. Here the node id is 1, the sequence is 0
// because it is the first id in that millisecond, and the decoded timestamp
// equals the input. This shows the timestamp round-trips exactly.
func ExampleNode_GenerateAt() {
	node, _ := snowflake.NewNode(1)
	id := node.GenerateAt(snowflake.Epoch + 1000)
	fmt.Println(snowflake.NodeOf(id))
	fmt.Println(snowflake.Sequence(id))
	fmt.Println(snowflake.Timestamp(id) == snowflake.Epoch+1000)
	// Output:
	// 1
	// 0
	// true
}

// ExampleNode_GenerateAt_sequence shows how monotonicity is preserved within a
// single millisecond. Two calls with the same timestamp cannot share an id, so
// the generator increments the 12-bit sequence field: the first id gets sequence
// 0 and the second gets sequence 1. This guarantees the two ids are distinct and
// strictly increasing even though their timestamps are identical. If the
// sequence were to overflow past 4095 the generator would advance to the next
// millisecond. Sequence numbering resets to zero whenever the clock moves to a
// new millisecond.
func ExampleNode_GenerateAt_sequence() {
	node, _ := snowflake.NewNode(0)
	ms := snowflake.Epoch + 2000
	a := node.GenerateAt(ms)
	b := node.GenerateAt(ms)
	fmt.Println(snowflake.Sequence(a), snowflake.Sequence(b))
	// Output: 0 1
}
