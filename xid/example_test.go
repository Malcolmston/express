package xid_test

import (
	"fmt"

	"github.com/malcolmston/express/xid"
)

// ExampleNew generates a sortable 12-byte identifier encoded as a 20-character
// string. Each id embeds a machine id, process id, and an atomically
// incrementing counter, so successive calls differ and the value cannot be
// printed deterministically. The example therefore checks the fixed
// 20-character length and confirms that Time extracts the same Unix-seconds
// timestamp that was supplied. The timestamp occupies the leading bytes, which
// is what makes xids sort chronologically. Here the decoded time matches the
// input second.
func ExampleNew() {
	id := xid.New(1600000000)
	t, err := xid.Time(id)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(id), t)
	// Output: 20 1600000000
}

// ExampleNewWithData builds an xid from every component explicitly, which makes
// the output fully deterministic. Passing the same timestamp, machine id, pid,
// and counter twice yields two identical strings, demonstrating that the
// construction is a pure function of its inputs. The result is still the canonical
// 20-character base32-hex form. This constructor is primarily useful for tests
// and reproducible output, whereas New drives the counter automatically. The
// equality check confirms determinism without depending on any random value.
func ExampleNewWithData() {
	a := xid.NewWithData(1600000000, [3]byte{1, 2, 3}, 100, 5)
	b := xid.NewWithData(1600000000, [3]byte{1, 2, 3}, 100, 5)
	fmt.Println(a == b, len(a))
	// Output: true 20
}
