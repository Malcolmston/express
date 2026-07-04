package snowflake

import "testing"

func TestNewNodeRange(t *testing.T) {
	if _, err := NewNode(-1); err == nil {
		t.Fatal("expected error for node id -1")
	}
	if _, err := NewNode(1024); err == nil {
		t.Fatal("expected error for node id 1024")
	}
	if _, err := NewNode(0); err != nil {
		t.Fatalf("node 0 rejected: %v", err)
	}
	if _, err := NewNode(1023); err != nil {
		t.Fatalf("node 1023 rejected: %v", err)
	}
}

func TestDecompose(t *testing.T) {
	const node = 512
	n, err := NewNode(node)
	if err != nil {
		t.Fatal(err)
	}
	const ms = int64(1700000000000)
	id := n.GenerateAt(ms)

	if NodeOf(id) != node {
		t.Errorf("NodeOf = %d, want %d", NodeOf(id), node)
	}
	if Timestamp(id) != ms {
		t.Errorf("Timestamp = %d, want %d", Timestamp(id), ms)
	}
	if Sequence(id) != 0 {
		t.Errorf("Sequence = %d, want 0", Sequence(id))
	}
}

func TestSameMillisecondSequence(t *testing.T) {
	n, err := NewNode(7)
	if err != nil {
		t.Fatal(err)
	}
	const ms = int64(1700000000000)
	id1 := n.GenerateAt(ms)
	id2 := n.GenerateAt(ms)

	if id1 == id2 {
		t.Fatal("ids for same ms are not distinct")
	}
	if Sequence(id2) <= Sequence(id1) {
		t.Errorf("sequence did not increase: %d then %d", Sequence(id1), Sequence(id2))
	}
	if id2 <= id1 {
		t.Errorf("id2 (%d) not greater than id1 (%d)", id2, id1)
	}
}

func TestIdsIncreaseWithTime(t *testing.T) {
	n, err := NewNode(3)
	if err != nil {
		t.Fatal(err)
	}
	id1 := n.GenerateAt(1700000000000)
	id2 := n.GenerateAt(1700000000001)
	if id2 <= id1 {
		t.Errorf("ids did not increase with time: %d then %d", id1, id2)
	}
	if Sequence(id2) != 0 {
		t.Errorf("sequence reset expected, got %d", Sequence(id2))
	}
}

func TestGenerateProducesValidNode(t *testing.T) {
	n, err := NewNode(42)
	if err != nil {
		t.Fatal(err)
	}
	id := n.Generate()
	if NodeOf(id) != 42 {
		t.Errorf("NodeOf(Generate) = %d, want 42", NodeOf(id))
	}
}
