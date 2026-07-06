package escapehtml_test

import (
	"fmt"

	"github.com/malcolmston/express/escapehtml"
)

// ExampleEscape rewrites the five HTML-significant characters into their entity
// equivalents so untrusted text can be embedded safely in markup. The ampersand
// becomes &amp;, the angle brackets become &lt; and &gt;, the double quote
// becomes &quot;, and the apostrophe becomes &#39;. Escaping both quote
// characters means the result is safe in element text and in attribute values
// alike. This is the baseline defense against HTML injection.
func ExampleEscape() {
	fmt.Println(escapehtml.Escape(`<a href="x">Tom & Jerry's</a>`))
	// Output: &lt;a href=&quot;x&quot;&gt;Tom &amp; Jerry&#39;s&lt;/a&gt;
}

// ExampleEscape_passthrough shows that text containing none of the five special
// characters is returned unchanged, with no allocation. Only &, <, >, ", and '
// are rewritten; letters, digits, punctuation, and multi-byte UTF-8 such as
// accented characters or emoji pass through verbatim. Here a plain sentence is
// returned exactly as given.
func ExampleEscape_passthrough() {
	fmt.Println(escapehtml.Escape("Café ☕ time"))
	// Output: Café ☕ time
}
