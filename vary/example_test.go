package vary_test

import (
	"fmt"
	"net/http"

	"github.com/malcolmston/express/vary"
)

// ExampleAppend adds field names to an existing Vary header value and returns
// the new value. Existing fields keep their order and casing, and appended
// fields are added in the order given, so the result is stable and predictable.
// Here a second field is appended to a header that already lists one. Append is
// the pure string form and also returns an error for invalid field names, which
// is nil here. This is what middleware uses to record that a response varies on
// a request header.
func ExampleAppend() {
	out, err := vary.Append("Accept-Encoding", "Accept-Language")
	fmt.Println(out, err)
	// Output: Accept-Encoding, Accept-Language <nil>
}

// ExampleAppend_dedup shows the case-insensitive deduplication. Because HTTP
// field names are case-insensitive, appending "accept-encoding" to a header that
// already contains "Accept-Encoding" leaves the value unchanged rather than
// listing the field twice. The original casing of the existing entry is
// preserved. This keeps the Vary header minimal and correct even when different
// middleware use different casings. No duplicate is ever introduced.
func ExampleAppend_dedup() {
	out, _ := vary.Append("Accept-Encoding", "accept-encoding")
	fmt.Println(out)
	// Output: Accept-Encoding
}

// ExampleVary mutates an http.Header in place, the form most convenient inside a
// handler. It reads the current Vary value, appends the given field, and writes
// the result back. Starting from an empty header, adding "Origin" sets the Vary
// header to exactly that field. Invalid field names are silently ignored so the
// call needs no error handling in a middleware chain. Here the header ends up
// carrying the single field.
func ExampleVary() {
	h := http.Header{}
	vary.Vary(h, "Origin")
	fmt.Println(h.Get("Vary"))
	// Output: Origin
}
