package kebabcase_test

import (
	"fmt"

	"github.com/malcolmston/express/kebabcase"
)

// This example shows the core behavior of KebabCase on a mixed-casing input.
// The space between the two words is turned into a hyphen, and the internal
// capital B in "fooBar" introduces an implicit word boundary because it
// follows a lowercase letter. Every letter is lower-cased as it is written.
// The result is the canonical kebab-case slug that the npm original would
// produce for the same string. This is the form typically used for URL slugs
// and CSS class names.
func ExampleKebabCase() {
	fmt.Println(kebabcase.KebabCase("fooBar Baz"))
	// Output: foo-bar-baz
}

// This example demonstrates how separators and consecutive capitals are
// handled. Underscores are treated as word separators just like spaces, so
// "foo_bar" splits into two words. A run of adjacent capitals such as the
// "XML" prefix in "XMLHttpRequest" is treated as a single word because no
// lowercase letter or digit precedes each capital. A boundary is only inserted
// where a lowercase letter is followed by a capital, which is why the split
// falls between "xmlhttp" and "request". These rules make the transformation
// deterministic on acronym-heavy identifiers.
func ExampleKebabCase_separators() {
	fmt.Println(kebabcase.KebabCase("foo_bar"))
	fmt.Println(kebabcase.KebabCase("XMLHttpRequest"))
	fmt.Println(kebabcase.KebabCase("foo123Bar"))
	// Output:
	// foo-bar
	// xmlhttp-request
	// foo123-bar
}

// This example covers the trimming and collapsing edge cases. Leading and
// trailing separators are stripped from the result, and any run of repeated
// separators collapses down to a single hyphen. Input that is empty or made up
// entirely of whitespace reduces to the empty string. This mirrors how the
// second normalization pass guarantees the output never begins, ends, or
// contains a doubled delimiter. The quoted printing makes the empty result
// visible in the output.
func ExampleKebabCase_trimming() {
	fmt.Printf("%q\n", kebabcase.KebabCase("__foo__bar__"))
	fmt.Printf("%q\n", kebabcase.KebabCase("--foo--bar--"))
	fmt.Printf("%q\n", kebabcase.KebabCase("   "))
	// Output:
	// "foo-bar"
	// "foo-bar"
	// ""
}
