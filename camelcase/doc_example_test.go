package camelcase_test

import (
	"fmt"

	"github.com/malcolmston/express/camelcase"
)

// ExampleCamelCase converts a delimiter-separated name into camelCase. Spaces,
// hyphens, underscores, and other non-alphanumeric characters all act as word
// boundaries and are discarded. The first word is lower-cased and every
// subsequent word is title-cased, so "foo-bar-baz" becomes "fooBarBaz". This is
// the transform used to turn config keys or header-derived names into
// program-friendly identifiers.
func ExampleCamelCase() {
	fmt.Println(camelcase.CamelCase("foo-bar-baz"))
	// Output: fooBarBaz
}

// ExamplePascalCase converts a name into PascalCase (also called
// UpperCamelCase), where every word including the first is title-cased. It
// shares the same word-splitting rules as CamelCase but upper-cases the leading
// character of the result. Here "foo-bar" becomes "FooBar". This form is handy
// for deriving exported-style type or constructor names.
func ExamplePascalCase() {
	fmt.Println(camelcase.PascalCase("foo-bar"))
	// Output: FooBar
}

// ExampleCamelCase_acronym demonstrates how runs of consecutive uppercase
// letters are handled. A case transition marks a word boundary, and a run of
// uppercase letters followed by a lowercase letter breaks just before that final
// letter, so "HTTPServer" splits into "HTTP" and "Server". Each word is then
// lower-cased and re-capitalized, collapsing the acronym, so the camelCase
// result is "httpServer".
func ExampleCamelCase_acronym() {
	fmt.Println(camelcase.CamelCase("HTTPServer"))
	// Output: httpServer
}
