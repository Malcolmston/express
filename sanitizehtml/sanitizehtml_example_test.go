package sanitizehtml_test

import (
	"fmt"

	"github.com/malcolmston/express/sanitizehtml"
)

// ExampleSanitize demonstrates cleaning an untrusted HTML fragment with the
// library defaults returned by DefaultOptions. The <p> and <b> tags are on the
// default allowlist, so they survive unchanged and keep their text. The
// <script> element is not on the allowlist, and because script contents are
// always discarded its "alert(1)" payload is removed entirely rather than being
// surfaced as text. The result is therefore safe to render back into a page,
// and the transformation is deterministic so an Output block can assert it.
func ExampleSanitize() {
	input := `<p>Hello <script>alert(1)</script><b>world</b></p>`
	fmt.Println(sanitizehtml.Sanitize(input, sanitizehtml.DefaultOptions()))
	// Output: <p>Hello <b>world</b></p>
}

// ExampleSanitize_attributes shows how the attribute allowlist filters
// individual attributes on an otherwise-allowed tag. The <a> tag is permitted
// by the defaults, and so is its href attribute, so both are kept. The onclick
// attribute is not in the permitted set for <a>, so it is dropped, which is the
// primary mechanism that blocks inline event-handler scripting. Retained
// attribute values are re-escaped on output, keeping the serialization
// consistent and safe.
func ExampleSanitize_attributes() {
	input := `<a href="/home" onclick="steal()">link</a>`
	fmt.Println(sanitizehtml.Sanitize(input, sanitizehtml.DefaultOptions()))
	// Output: <a href="/home">link</a>
}

// ExampleSanitize_customPolicy demonstrates a restrictive custom policy that
// allows only the <em> tag and no attributes at all. The <em> tag is kept, but
// the <span> tag is not on the allowlist and is stripped while its inner text
// "keep" is preserved. This illustrates the text-preserving removal of
// disallowed tags: markup disappears but the reader's words remain. Building
// Options directly gives full control over the tag and attribute policy.
func ExampleSanitize_customPolicy() {
	opts := sanitizehtml.Options{AllowedTags: []string{"em"}}
	input := `<span>keep</span> <em>this</em>`
	fmt.Println(sanitizehtml.Sanitize(input, opts))
	// Output: keep <em>this</em>
}
