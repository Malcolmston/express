package typeis_test

import (
	"fmt"

	"github.com/malcolmston/express/typeis"
)

// ExampleIs tests a concrete Content-Type against a candidate. The candidate
// "json" is an extension-style shorthand that expands to "application/json" for
// matching. The parameters on the actual header (here "; charset=utf-8") are
// stripped during normalization, so they never affect the result. Following the
// upstream type-is convention, a matching non-wildcard candidate is returned
// exactly as supplied (here "json") together with true; on no match Is returns
// "" and false. This is the same convenience Express exposes as req.is.
func ExampleIs() {
	match, ok := typeis.Is("application/json; charset=utf-8", "json")
	fmt.Println(match, ok)
	// Output: json true
}

// ExampleIs_multiple tries several candidates in order and reports the first one
// that matches. Here the value is HTML, so the "json" candidate fails and the
// "html" candidate matches. As with upstream type-is, a matching non-wildcard
// candidate is echoed back as supplied (here "html"). Candidates may freely mix
// shorthands, full types, wildcards, and suffix forms; the first successful
// match wins. This makes it easy to branch on the several content types a
// handler is willing to accept.
func ExampleIs_multiple() {
	match, ok := typeis.Is("text/html", "json", "html")
	fmt.Println(match, ok)
	// Output: html true
}

// ExampleMatch compares an already-expanded pattern against a concrete type
// directly, with wildcard and suffix support. "application/*" matches any
// application subtype, and the suffix pattern "*/*+json" (the expanded form of
// the "+json" shorthand) matches "application/vnd.api+json" because the two
// suffixes agree. A pattern whose type does not match, such as "text/*" against
// an application type, returns false. Both arguments are normalized, so
// parameters are ignored. Match is the lower-level primitive that Is builds on
// after expanding shorthands.
func ExampleMatch() {
	fmt.Println(typeis.Match("application/*", "application/json"))
	fmt.Println(typeis.Match("*/*+json", "application/vnd.api+json"))
	fmt.Println(typeis.Match("text/*", "application/json"))
	// Output:
	// true
	// true
	// false
}
