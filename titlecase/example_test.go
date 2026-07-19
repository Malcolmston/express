package titlecase_test

import (
	"fmt"

	"github.com/malcolmston/express/titlecase"
)

// ExampleTitleCase capitalizes the significant words of a plain phrase while
// leaving small function words such as "of" and "the" in lower case
// mid-sentence, matching the upstream change-case title-case library. Here
// "the quick brown fox jumps over the lazy dog" keeps "over" and the interior
// "the" lower case.
func ExampleTitleCase() {
	fmt.Println(titlecase.TitleCase("the quick brown fox jumps over the lazy dog"))
	// Output: The Quick Brown Fox Jumps over the Lazy Dog
}

// ExampleTitleCase_preserved shows that the transformation never lower-cases
// letters: acronyms like "NASA" and manually cased words like "camelCase" pass
// through unchanged, and only the first letter of each capitalized word is
// upper-cased.
func ExampleTitleCase_preserved() {
	fmt.Println(titlecase.TitleCase("we keep NASA capitalized"))
	fmt.Println(titlecase.TitleCase("pass camelCase through"))
	// Output:
	// We Keep NASA Capitalized
	// Pass camelCase Through
}

// ExampleTitleCase_sentenceCase demonstrates the SentenceCase option, which
// capitalizes only the first word of each sentence rather than every
// significant word.
func ExampleTitleCase_sentenceCase() {
	fmt.Println(titlecase.TitleCase("the iPhone: a quote", titlecase.Options{SentenceCase: true}))
	// Output: The iPhone: a quote
}
