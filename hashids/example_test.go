package hashids_test

import (
	"fmt"
	"log"

	"github.com/malcolmston/express/hashids"
)

// ExampleHashID_Encode shows how to turn a single non-negative integer into a
// short, non-sequential hash. A codec is created with a salt, which diversifies
// the output so that the same numbers produce different hashes under different
// salts. The minimum length of 0 means no padding is added and the hash is as
// short as the algorithm allows. Encoding is deterministic: the same salt,
// alphabet and numbers always yield the same string. Here the salt
// "this is my salt" encodes 12345 to the canonical Hashids value "NkK9".
func ExampleHashID_Encode() {
	h, err := hashids.New("this is my salt", 0)
	if err != nil {
		log.Fatal(err)
	}
	hash, err := h.Encode(12345)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hash)
	// Output: NkK9
}

// ExampleHashID_Decode demonstrates the inverse of Encode, recovering the
// original integers from a hash. Decode uses the same salt and alphabet that
// produced the hash and validates its work by re-encoding the result. Multiple
// numbers can be encoded together and are returned in the same order. This
// example round-trips the slice {1, 2, 3}, which encodes to "laHquq" under the
// sample salt, back to its original values. A hash that does not decode
// consistently would instead yield an empty slice.
func ExampleHashID_Decode() {
	h, err := hashids.New("this is my salt", 0)
	if err != nil {
		log.Fatal(err)
	}
	hash, err := h.Encode(1, 2, 3)
	if err != nil {
		log.Fatal(err)
	}
	nums, err := h.Decode(hash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hash)
	fmt.Println(nums)
	// Output:
	// laHquq
	// [1 2 3]
}
