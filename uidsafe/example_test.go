package uidsafe_test

import (
	"fmt"

	"github.com/malcolmston/express/uidsafe"
)

// ExampleBytes generates a URL-safe token from cryptographically random bytes.
// The value is random and cannot be printed deterministically, so the example
// checks its length instead. Because base64 packs three input bytes into four
// output characters, 18 random bytes produce a 24-character string with no
// padding. The token uses the URL-safe alphabet, so it can be dropped into a
// path, query value, or cookie without escaping. Bytes returns an error only if
// the system random source fails.
func ExampleBytes() {
	s, err := uidsafe.Bytes(18)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(s))
	// Output: 24
}

// ExampleMustBytes is the panicking convenience form, suitable for program
// initialization where a failing random source is unrecoverable. As with Bytes
// the output length follows the base64 expansion: 15 random bytes yield a
// 20-character string. The result is again random, so only the length is
// asserted here. MustBytes never returns an error; it panics instead. This makes
// it convenient for package-level variables that must be seeded at startup.
func ExampleMustBytes() {
	fmt.Println(len(uidsafe.MustBytes(15)))
	// Output: 20
}
