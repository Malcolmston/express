package globtoregexp_test

import (
	"fmt"

	"github.com/malcolmston/express/globtoregexp"
)

// ExampleGlobToRegExp converts a glob using the package defaults, where both
// Extended and GlobStar are off. In this mode a run of "*" becomes ".*", which
// matches any characters including path separators. The generated pattern is
// anchored with "^" and "$" so it must match the entire input. Here "*.js"
// matches both a plain filename and one nested under a directory, since the
// default star crosses separators. The compiled pattern string is also shown.
func ExampleGlobToRegExp() {
	re, err := globtoregexp.GlobToRegExp("*.js")
	if err != nil {
		panic(err)
	}
	fmt.Println(re.String())
	fmt.Println(re.MatchString("foo.js"))
	fmt.Println(re.MatchString("foo/bar.js"))
	// Output:
	// ^.*\.js$
	// true
	// true
}

// ExampleGlobToRegExpOpts_globStar enables GlobStar so that a single "*" no
// longer crosses path separators. With this option a single star matches within
// one path segment and compiles to "([^/]*)". As a result "*.js" matches
// "foo.js" but not "foo/bar.js", because the latter contains a separator. This
// behavior is closer to conventional shell globbing. The generated pattern is
// printed alongside the match results.
func ExampleGlobToRegExpOpts_globStar() {
	re, err := globtoregexp.GlobToRegExpOpts("*.js", globtoregexp.Options{GlobStar: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(re.String())
	fmt.Println(re.MatchString("foo.js"))
	fmt.Println(re.MatchString("foo/bar.js"))
	// Output:
	// ^([^/]*)\.js$
	// true
	// false
}

// ExampleGlobToRegExpOpts_globStarSegments shows the "**" globstar spanning
// multiple path segments. With GlobStar on, a "**" bounded by separators or the
// ends of the pattern matches zero or more whole segments. The pattern "a/**/b"
// therefore matches "a/b", "a/x/b", and "a/x/y/b" alike. This is useful for
// recursive directory matching. Each of the three inputs below matches.
func ExampleGlobToRegExpOpts_globStarSegments() {
	re, err := globtoregexp.GlobToRegExpOpts("a/**/b", globtoregexp.Options{GlobStar: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(re.MatchString("a/b"))
	fmt.Println(re.MatchString("a/x/b"))
	fmt.Println(re.MatchString("a/x/y/b"))
	// Output:
	// true
	// true
	// true
}

// ExampleGlobToRegExpOpts_extended turns on Extended mode, which activates extra
// glob syntax. In this mode "?" matches exactly one character, "[...]" character
// classes pass through to the regexp, and "{a,b}" groups become alternations.
// Here "*.{js,ts}" accepts both ".js" and ".ts" extensions but rejects ".md".
// Without Extended those brace and bracket characters would be treated
// literally. GlobStar is also enabled so "*" stays within a single segment.
func ExampleGlobToRegExpOpts_extended() {
	re, err := globtoregexp.GlobToRegExpOpts("*.{js,ts}", globtoregexp.Options{Extended: true, GlobStar: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(re.MatchString("foo.js"))
	fmt.Println(re.MatchString("foo.ts"))
	fmt.Println(re.MatchString("foo.md"))
	// Output:
	// true
	// true
	// false
}

// ExampleGlobToRegExpOpts_flags demonstrates the JavaScript-style RegExp flags.
// The "i" flag makes matching case-insensitive by prefixing the pattern with
// "(?i)", so a glob written in uppercase still matches lowercase input. The "g"
// flag disables the "^...$" anchoring, allowing the pattern to match a substring
// rather than the whole string. Below, "*.JS" with "i" matches "foo.js", and
// "bar" with "g" matches inside "foobarbaz". Both flags can be combined in the
// same Flags string.
func ExampleGlobToRegExpOpts_flags() {
	re, _ := globtoregexp.GlobToRegExpOpts("*.JS", globtoregexp.Options{GlobStar: true, Flags: "i"})
	fmt.Println(re.MatchString("foo.js"))

	re2, _ := globtoregexp.GlobToRegExpOpts("bar", globtoregexp.Options{Flags: "g"})
	fmt.Println(re2.String())
	fmt.Println(re2.MatchString("foobarbaz"))
	// Output:
	// true
	// bar
	// true
}
