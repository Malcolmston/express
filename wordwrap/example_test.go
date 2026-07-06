package wordwrap_test

import (
	"fmt"

	"github.com/malcolmston/express/wordwrap"
)

// ExampleWrap reflows a sentence so that no line exceeds the chosen width. With
// a Width of 20 and no indent, words are packed greedily onto each line until
// the next word would push it past the limit, at which point a new line begins.
// Width is measured in runes and counts only the text, excluding any indent.
// Existing newlines would be treated as hard paragraph breaks, but this input
// has none. The result is three wrapped lines joined by the default "\n".
func ExampleWrap() {
	text := "The quick brown fox jumps over the lazy dog"
	fmt.Println(wordwrap.Wrap(text, wordwrap.Options{Width: 20}))
	// Output:
	// The quick brown fox
	// jumps over the lazy
	// dog
}

// ExampleWrap_cut demonstrates the Cut option, which breaks words that are
// longer than the width instead of letting them overflow. Here the eight-letter
// run is chopped into four-character pieces before wrapping, so the hard width
// limit is never exceeded. Without Cut such a word would be emitted whole on its
// own line. The shorter trailing word then wraps normally. This is useful when a
// strict column limit must be honored even for unbroken tokens.
func ExampleWrap_cut() {
	fmt.Println(wordwrap.Wrap("aaaaaaaa bb", wordwrap.Options{Width: 4, Cut: true}))
	// Output:
	// aaaa
	// aaaa
	// bb
}
