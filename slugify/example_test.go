package slugify_test

import (
	"fmt"

	"github.com/malcolmston/express/slugify"
)

// ExampleSlugify converts a title into a URL-safe slug using the library
// defaults. Called with no Options it enables trimming, uses "-" as the
// separator, and does not lowercase, matching the npm defaults. Punctuation
// that is neither a word character nor whitespace is removed, whitespace runs
// collapse into a single separator, and no doubled separator survives. Here the
// comma and exclamation mark are dropped and the space becomes a hyphen. Note
// that casing is preserved because lowercasing is off by default.
func ExampleSlugify() {
	fmt.Println(slugify.Slugify("Hello, World!"))
	// Output: Hello-World
}

// ExampleSlugify_transliterate shows Unicode transliteration together with the
// Lower option. Accented Latin letters are mapped to their ASCII equivalents
// through the built-in character table, so "è" and "û" and "é" become "e", "u",
// and "e". With Lower set to true the final slug is lowercased. An explicit
// Options value is honored verbatim except that an empty Separator falls back to
// "-". The result is a clean, path-safe ASCII string.
func ExampleSlugify_transliterate() {
	fmt.Println(slugify.Slugify("Crème Brûlée", slugify.Options{Lower: true}))
	// Output: creme-brulee
}

// ExampleSlugify_symbols demonstrates symbol replacement, where certain
// characters are transliterated to words rather than removed. The ampersand is
// mapped to "and" before separators are applied, so "Me & You" becomes a
// readable slug rather than losing the conjunction. Surrounding whitespace is
// collapsed and joined with the default hyphen separator. This mirrors the npm
// library's handling of common symbols. Symbols without a mapping are stripped
// instead.
func ExampleSlugify_symbols() {
	fmt.Println(slugify.Slugify("Me & You"))
	// Output: Me-and-You
}
