package ulid_test

import (
	"fmt"

	"github.com/malcolmston/express/ulid"
)

// ExampleNewWithEntropy builds a ULID deterministically from an explicit
// timestamp and caller-supplied entropy, which makes the output reproducible.
// With a timestamp of zero and ten zero bytes of entropy every bit is zero, so
// the Crockford base32 encoding renders 26 '0' characters. This shows the fixed
// 26-character canonical length and the encoding directly. In real use the
// entropy would come from crypto/rand via New. Supplying entropy by hand is
// primarily useful for tests and reproducible generation.
func ExampleNewWithEntropy() {
	id, err := ulid.NewWithEntropy(0, make([]byte, 10))
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	fmt.Println(len(id))
	// Output:
	// 00000000000000000000000000
	// 26
}

// ExampleNew generates a ULID for a given millisecond timestamp using random
// entropy. Because the entropy is random the string cannot be printed
// deterministically, so the example verifies its length and that the embedded
// timestamp round-trips through Timestamp. The high 48 bits hold the millisecond
// timestamp and the low 80 bits hold entropy, and Timestamp extracts just the
// time field without a full decode. Since the timestamp occupies the most
// significant bits, ULIDs sort chronologically as strings. Here the decoded time
// matches the value passed in.
func ExampleNew() {
	id, err := ulid.New(1234567890123)
	if err != nil {
		panic(err)
	}
	ts, _ := ulid.Timestamp(id)
	fmt.Println(len(id), ts == 1234567890123)
	// Output: 26 true
}
