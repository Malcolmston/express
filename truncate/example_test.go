package truncate_test

import (
	"fmt"

	"github.com/malcolmston/express/truncate"
)

// ExampleTruncate shortens a string to at most the given number of runes,
// appending the default ellipsis "..." when content is cut. The length budget is
// inclusive of the ellipsis, so with a limit of 8 only five characters of
// content are kept and the three-character ellipsis fills the rest, never
// exceeding the requested length. Length is measured in runes, so a multibyte
// character is never split. A string that already fits within the limit is
// returned unchanged with no ellipsis.
func ExampleTruncate() {
	fmt.Println(truncate.Truncate("Hello, World!", 8))
	// Output: Hello...
}

// ExampleTruncateOpts_wordBoundary trims the cut content back to the last
// whitespace so a word is not sliced in half. With a limit of 14 and the default
// ellipsis, the raw cut would land inside "brown", but WordBoundary rolls it back
// to the previous space, yielding "The quick" plus the ellipsis. As always the
// ellipsis length counts toward the total. This produces cleaner previews where
// the final visible word is whole. If there were no whitespace in the kept span
// the content would be left as-is.
func ExampleTruncateOpts_wordBoundary() {
	out := truncate.TruncateOpts("The quick brown fox", 14, truncate.Options{
		Ellipsis:     "...",
		WordBoundary: true,
	})
	fmt.Println(out)
	// Output: The quick...
}
