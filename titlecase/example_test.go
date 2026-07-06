package titlecase_test

import (
	"fmt"

	"github.com/malcolmston/express/titlecase"
)

// ExampleTitleCase capitalizes the first letter of each word in a plain phrase.
// The input is split into words on whitespace, each word is recased so its first
// letter is uppercase and the rest lowercase, and the words are rejoined with a
// single space. Here "hello world" becomes "Hello World". Every word is
// capitalized, since this package does not implement small-word handling for
// function words like "of" or "the". The transformation is deterministic and
// depends only on the input.
func ExampleTitleCase() {
	fmt.Println(titlecase.TitleCase("hello world"))
	// Output: Hello World
}

// ExampleTitleCase_boundaries shows the word-boundary detection that makes the
// function robust across casing and separator styles. A camelCase boundary in
// "helloWorld" splits into "hello" and "World", and runs of punctuation or
// underscores and hyphens in "foo_bar-baz" separate three words. Each detected
// word is then capitalized independently and joined with single spaces, so mixed
// or repeated separators collapse to one space. This is why identifiers from
// code convert cleanly into human-friendly display strings.
func ExampleTitleCase_boundaries() {
	fmt.Println(titlecase.TitleCase("helloWorld"))
	fmt.Println(titlecase.TitleCase("foo_bar-baz"))
	// Output:
	// Hello World
	// Foo Bar Baz
}
