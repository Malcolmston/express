package statuses_test

import (
	"fmt"

	"github.com/malcolmston/express/statuses"
)

// ExampleMessage looks up the canonical reason phrase for a numeric HTTP status
// code. The Message function performs a direct table lookup and returns the
// registered phrase, such as "Not Found" for 404 and "Internal Server Error" for
// 500. When the code is not part of the known table it returns the empty string
// rather than an error, which makes it convenient for building default response
// text. This example prints the phrase for a couple of well-known codes and then
// shows the empty result for an unregistered code. The output is fully
// deterministic because the underlying table is fixed.
func ExampleMessage() {
	fmt.Printf("%q\n", statuses.Message(404))
	fmt.Printf("%q\n", statuses.Message(500))
	fmt.Printf("%q\n", statuses.Message(799))
	// Output:
	// "Not Found"
	// "Internal Server Error"
	// ""
}

// ExampleCode performs the reverse lookup, turning a reason phrase back into its
// numeric status code. Matching is case-insensitive and tolerant of surrounding
// whitespace because Code trims and lower-cases its argument before consulting
// the table. A recognized phrase yields the code and a nil error, so both
// "Not Found" and "not found" resolve to 404. An unrecognized phrase yields a
// zero code together with a non-nil error, which this example prints to show the
// failure path.
func ExampleCode() {
	code, err := statuses.Code("Not Found")
	fmt.Println(code, err)

	code, err = statuses.Code("not found")
	fmt.Println(code, err)

	code, err = statuses.Code("nonsense")
	fmt.Println(code, err)
	// Output:
	// 404 <nil>
	// 404 <nil>
	// 0 invalid status message: "nonsense"
}

// ExampleIsRedirect demonstrates classifying status codes by behavior. The three
// predicate functions each consult a fixed set of codes: IsRedirect covers the
// 3xx codes that carry a Location header, IsRetry covers the gateway-family codes
// that a client may safely retry, and IsEmpty covers the codes whose responses
// must not include a body. This example checks a redirect (302), confirms that
// 304 is not treated as a redirect, checks a retriable gateway timeout (504),
// and checks an empty-body code (204). The boolean results are deterministic and
// mirror the classification tables of the npm original.
func ExampleIsRedirect() {
	fmt.Println("302 redirect:", statuses.IsRedirect(302))
	fmt.Println("304 redirect:", statuses.IsRedirect(304))
	fmt.Println("504 retry:", statuses.IsRetry(504))
	fmt.Println("204 empty:", statuses.IsEmpty(204))
	// Output:
	// 302 redirect: true
	// 304 redirect: false
	// 504 retry: true
	// 204 empty: true
}

// ExampleCodes lists every known status code in ascending order and pairs each
// with its reason phrase. Codes returns a freshly sorted slice, so iterating it
// visits the table in a stable, predictable sequence. This example prints only
// the first few entries to keep the output short while still demonstrating both
// the ordering and the round trip between a code and its Message. The values are
// deterministic because the backing table never changes at runtime. Such
// iteration is handy for validating that Message and Code agree for every code.
func ExampleCodes() {
	for _, code := range statuses.Codes()[:3] {
		fmt.Printf("%d %s\n", code, statuses.Message(code))
	}
	// Output:
	// 100 Continue
	// 101 Switching Protocols
	// 102 Processing
}
