package cryptorandomstring_test

import (
	"fmt"

	"github.com/malcolmston/express/cryptorandomstring"
)

// ExampleHex generates a cryptographically strong random hex string of the
// requested length. Because the output is random it cannot be shown literally,
// so this example prints its length instead, which is deterministic. Length is
// measured in characters (runes), so a request for 16 yields exactly 16 hex
// digits. Hex is the convenient default for tokens and identifiers that only
// need the 0-9a-f alphabet.
func ExampleHex() {
	s, err := cryptorandomstring.Hex(16)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(len(s))
	// Output: 16
}

// ExampleGenerate produces a random string from a chosen character-set preset.
// The Options struct selects the length and the type; here the "numeric" preset
// restricts the output to the digits 0-9. Sampling is unbiased because it draws
// each character with crypto/rand.Int rather than reducing a byte modulo the set
// size. The example verifies the length and that every character is a digit,
// both of which are deterministic.
func ExampleGenerate() {
	s, err := cryptorandomstring.Generate(cryptorandomstring.Options{
		Length: 8,
		Type:   "numeric",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	allDigits := true
	for _, r := range s {
		if r < '0' || r > '9' {
			allDigits = false
		}
	}
	fmt.Println(len(s), allDigits)
	// Output: 8 true
}
