package shortid_test

import (
	"fmt"

	"github.com/malcolmston/express/shortid"
)

// ExampleGenerate produces a short, URL-friendly identifier. Because the id
// embeds the current time and six random characters it is different on every
// call, so the example validates the result with IsValid rather than printing
// the random value. IsValid confirms the id is non-empty and composed entirely
// of characters from the current 64-symbol alphabet. Ids are typically 7 to 14
// characters long depending on the magnitude of the timestamp. Every character
// is URL-safe, so the id never needs percent-encoding.
func ExampleGenerate() {
	id, err := shortid.Generate()
	if err != nil {
		panic(err)
	}
	fmt.Println(shortid.IsValid(id))
	// Output: true
}

// ExampleIsValid checks whether a string could have been produced by this
// package, meaning it is non-empty and contains only characters from the
// current alphabet. The default alphabet is the URL-safe set of letters,
// digits, "_" and "-", so a value using those characters validates. An empty
// string is rejected, and a string containing a space fails because a space is
// not in the alphabet. This is a cheap membership check, not a cryptographic
// verification of provenance.
func ExampleIsValid() {
	fmt.Println(shortid.IsValid("abc-_XYZ"))
	fmt.Println(shortid.IsValid(""))
	fmt.Println(shortid.IsValid("has space"))
	// Output:
	// true
	// false
	// false
}
