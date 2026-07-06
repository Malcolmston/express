package randomstring_test

import (
	"fmt"
	"strings"

	"github.com/malcolmston/express/randomstring"
)

// ExampleGenerate draws a random string of a requested length from a named
// preset character set. Because the output is random it cannot be printed
// verbatim in a deterministic example, so this checks the two guaranteed
// properties instead: the length matches what was requested, and every
// character comes from the chosen "numeric" preset. Sampling uses crypto/rand
// for unbiased, security-grade randomness. An empty charset name would default
// to "alphanumeric". Here a 16-character numeric string is produced.
func ExampleGenerate() {
	s, err := randomstring.Generate(16, "numeric")
	if err != nil {
		panic(err)
	}
	allDigits := strings.Trim(s, "0123456789") == ""
	fmt.Println(len(s), allDigits)
	// Output: 16 true
}

// ExampleNew returns the library's default token: a 32-character alphanumeric
// string suitable for identifiers and one-off secrets. The value is random, so
// the example verifies only its length rather than its contents. New is a thin
// convenience wrapper over Generate with the "alphanumeric" preset and a length
// of 32. Every character is drawn from crypto/rand, so the result is safe for
// security-sensitive use. This is the typical starting point when you just need
// a random token.
func ExampleNew() {
	s, err := randomstring.New()
	if err != nil {
		panic(err)
	}
	fmt.Println(len(s))
	// Output: 32
}
