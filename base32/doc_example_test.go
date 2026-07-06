package base32_test

import (
	"fmt"

	"github.com/malcolmston/express/base32"
)

// ExampleEncode encodes raw bytes into the canonical padded RFC 4648 base32
// form. The input is treated as a bit stream, split into 5-bit groups, and each
// group is mapped to one of A-Z or 2-7. Because five bits do not divide evenly
// into bytes, the output is padded with "=" so its length is a multiple of
// eight. Base32 is handy for identifiers that must survive case-insensitive
// systems or be typed by hand.
func ExampleEncode() {
	fmt.Println(base32.Encode([]byte("hello")))
	// Output: NBSWY3DP
}

// ExampleEncodeNoPadding produces the same base32 characters as Encode but drops
// the trailing "=" padding run. This is convenient for contexts such as URLs or
// filenames where the "=" character is unwanted or reserved. The unpadded form
// still decodes cleanly because Decode tolerates missing padding. Here the
// two-byte input "hi" encodes to four characters with no padding.
func ExampleEncodeNoPadding() {
	fmt.Println(base32.EncodeNoPadding([]byte("hi")))
	// Output: NBUQ
}

// ExampleDecode reverses base32 encoding and is deliberately lenient. Before
// decoding it trims surrounding whitespace, upper-cases the text so lowercase
// input is accepted, and strips any trailing "=" so both padded and unpadded
// strings work. An invalid alphabet character or length yields a non-nil error
// instead of partial output. Here a lowercase, unpadded string round-trips back
// to the original bytes.
func ExampleDecode() {
	data, err := base32.Decode("nbuq")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%s\n", data)
	// Output: hi
}
