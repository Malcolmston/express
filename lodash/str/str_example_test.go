package str_test

import (
	"fmt"

	"github.com/malcolmston/express/lodash/str"
)

// ExampleWords shows the word splitter that underpins the case converters. It
// scans the input and emits maximal runs of letters or digits, treating every
// other rune as a separator. It also splits on case boundaries, so camelCase
// input decomposes into its component words. A run of upper-case letters is
// kept together as an acronym, except that the final upper-case letter attaches
// to a following lower-case word, which is why "XMLHttpTest" splits into three
// words. The %q verb quotes each element so the boundaries are visible.
func ExampleWords() {
	fmt.Printf("%q\n", str.Words("fooBar"))
	fmt.Printf("%q\n", str.Words("XMLHttpTest"))
	// Output:
	// ["foo" "Bar"]
	// ["XML" "Http" "Test"]
}

// ExampleCamelCase converts arbitrary text to camel case. It first splits the
// input into words, then lower-cases each word and upper-cases the first letter
// of every word after the first, joining them with no separator. Leading,
// trailing and repeated separators are collapsed, so surrounding dashes or
// underscores disappear. Fully upper-cased input is normalized down to camel
// case as well. All three inputs here therefore produce the same result.
func ExampleCamelCase() {
	fmt.Println(str.CamelCase("Foo Bar"))
	fmt.Println(str.CamelCase("--foo-bar--"))
	fmt.Println(str.CamelCase("__FOO_BAR__"))
	// Output:
	// fooBar
	// fooBar
	// fooBar
}

// ExampleKebabCase converts text to kebab case, lower-casing each word and
// joining them with hyphens. It is a common way to derive URL slugs or CSS
// class names from a label. Because it is built on the same word splitter as
// the other converters, it also breaks apart camelCase and acronym-heavy
// identifiers. Here "XMLHttpRequest" becomes three lower-cased, hyphen-joined
// words. Separators in the input are normalized away.
func ExampleKebabCase() {
	fmt.Println(str.KebabCase("fooBar"))
	fmt.Println(str.KebabCase("XMLHttpRequest"))
	// Output:
	// foo-bar
	// xml-http-request
}

// ExampleSnakeCase converts text to snake case, lower-casing each word and
// joining them with underscores. Like the other converters it splits camelCase,
// kebab-case and spaced input uniformly. This form is common for database
// column names and environment-variable-like keys. Leading and trailing
// separators in the input are discarded. The example converts a spaced label.
func ExampleSnakeCase() {
	fmt.Println(str.SnakeCase("Foo Bar"))
	fmt.Println(str.SnakeCase("fooBar"))
	// Output:
	// foo_bar
	// foo_bar
}

// ExampleStartCase converts text to start case, upper-casing the first letter
// of every word and joining the words with single spaces. Unlike Capitalize it
// preserves the remaining letters of each word rather than lower-casing them,
// which is why an acronym like "XML" stays upper-cased. It is well suited to
// producing human-readable titles from identifiers. Surrounding separators are
// removed during splitting. The two examples show a slug and an acronym.
func ExampleStartCase() {
	fmt.Println(str.StartCase("--foo-bar--"))
	fmt.Println(str.StartCase("XMLHttp"))
	// Output:
	// Foo Bar
	// XML Http
}

// ExamplePad centers a string within a target width by adding padding on both
// sides. When the string is already at least as long as the target it is
// returned unchanged. The padding characters are taken from the chars argument
// and repeated, then truncated so the total length is exactly right. When the
// two sides cannot be split evenly the extra padding goes on the right. Here
// "abc" is padded to width 8 using the two-character pattern "_-".
func ExamplePad() {
	fmt.Println(str.Pad("abc", 8, "_-"))
	// Output:
	// _-abc_-_
}

// ExampleTruncate shortens a string that exceeds a maximum length, replacing
// the tail with an omission marker. With a zero-value options struct the length
// defaults to 30 and the omission defaults to "...". The omission is counted
// against the length budget, so the retained text plus the marker fit within
// the limit. Strings already within the limit are returned unchanged. This
// example uses the defaults on a long string.
func ExampleTruncate() {
	fmt.Println(str.Truncate("hi-diddly-ho there, neighborino", str.TruncateOptions{}))
	// Output:
	// hi-diddly-ho there, neighbo...
}

// ExampleDeburr folds accented Latin letters to their basic ASCII equivalents
// and removes combining diacritical marks. It covers the Latin-1 Supplement and
// the more common Latin Extended-A letters, which is enough for most Western
// European text. This is useful for normalizing text before case-insensitive
// search or slug generation. Characters outside the mapped set pass through
// unchanged. Here an accented French phrase is reduced to plain ASCII.
func ExampleDeburr() {
	fmt.Println(str.Deburr("déjà vu"))
	// Output:
	// deja vu
}

// ExampleEscape converts the five HTML-significant characters (&, <, >, " and
// ') into their corresponding HTML entities. It is meant for interpolating
// untrusted text into HTML content, not for general-purpose sanitization. Only
// those five characters are affected; all other text is left as is. The inverse
// operation is provided by Unescape. Here an ampersand is converted to &amp;.
func ExampleEscape() {
	fmt.Println(str.Escape("fred, barney, & pebbles"))
	// Output:
	// fred, barney, &amp; pebbles
}

// ExampleParseInt parses a string into an integer, mirroring JavaScript's
// parseInt. A radix of 10 reads a decimal number, tolerating a leading zero.
// A radix of 0 auto-detects a "0x" prefix and parses as hexadecimal. Parsing
// stops at the first character that is not a valid digit for the radix, so
// trailing text like a unit suffix is ignored rather than causing an error.
// The three cases show decimal, hex auto-detection and trailing-text handling.
func ExampleParseInt() {
	fmt.Println(str.ParseInt("08", 10))
	fmt.Println(str.ParseInt("0x1A", 0))
	fmt.Println(str.ParseInt("42px", 10))
	// Output:
	// 8
	// 26
	// 42
}
