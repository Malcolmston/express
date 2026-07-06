package nanoid_test

import (
	"fmt"
	"strings"

	"github.com/malcolmston/express/nanoid"
)

// ExampleNew generates an identifier with the package defaults: the 64-character
// URL-safe alphabet and a length of 21. Because the bytes come from crypto/rand
// the value is different on every run, so the example asserts on its properties
// rather than printing the id itself. It confirms the id is exactly DefaultSize
// characters long and that every character is drawn from DefaultAlphabet, which
// is what makes the id safe to drop into a path or query string without escaping.
// The strings.Trim call removes all alphabet characters; an empty remainder means
// no character fell outside the alphabet.
func ExampleNew() {
	id, err := nanoid.New()
	if err != nil {
		panic(err)
	}
	fmt.Println(len(id) == nanoid.DefaultSize)
	fmt.Println(strings.Trim(id, nanoid.DefaultAlphabet) == "")
	// Output:
	// true
	// true
}

// ExampleNewSize keeps the default URL-safe alphabet but chooses a custom length,
// here 10 characters. A shorter id has a smaller keyspace and therefore a higher
// collision probability, so the length is a deliberate trade-off the caller makes
// against compactness. As with New the value is random, so the example prints the
// length rather than the id. This is the right entry point when the default
// 21-character id is longer than a use case needs.
func ExampleNewSize() {
	id, err := nanoid.NewSize(10)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(id))
	// Output: 10
}

// ExampleCustom supplies both a custom alphabet and a custom size, restricting
// the id to lowercase hexadecimal digits. The unbiased mask-and-reject sampler
// ensures every character of the alphabet is equally likely despite the alphabet
// length not being a power of two. The example verifies the length is 16 and that
// every character is a valid hex digit. Custom is the entry point for domain
// constraints such as case-insensitive tokens or a reduced character set.
func ExampleCustom() {
	id, err := nanoid.Custom("0123456789abcdef", 16)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(id))
	fmt.Println(strings.Trim(id, "0123456789abcdef") == "")
	// Output:
	// 16
	// true
}
