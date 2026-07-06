package urljoin_test

import (
	"fmt"

	"github.com/malcolmston/express/urljoin"
)

// ExampleURLJoin assembles a URL from a base and path fragments, inserting
// exactly one slash between each part. The scheme's "://" separator is preserved
// rather than collapsed, so the protocol stays intact while the path segments
// are joined cleanly. This is the common case of building an endpoint URL from
// pieces without worrying about which piece carried a slash. The result is a
// single well-formed URL. Empty parts, if any, would simply be skipped.
func ExampleURLJoin() {
	fmt.Println(urljoin.URLJoin("http://example.com", "foo", "bar"))
	// Output: http://example.com/foo/bar
}

// ExampleURLJoin_slashes shows the slash normalization that gives the package
// its purpose. Even though several parts carry leading or trailing slashes, the
// duplicates are collapsed so the output never contains a doubled slash between
// segments. Leading slashes are stripped from every part after the first, and
// trailing slashes from every part before the last. The protocol's own double
// slash is left untouched. The result is identical to the clean join above.
func ExampleURLJoin_slashes() {
	fmt.Println(urljoin.URLJoin("http://example.com/", "/foo/", "/bar"))
	// Output: http://example.com/foo/bar
}

// ExampleURLJoin_query combines query strings from multiple parts into one
// well-formed query. The first "?" is kept as the query introducer and every
// subsequent "?" is rewritten to "&", so two parts each carrying a query merge
// into a single "?a=1&b=2". A slash that would otherwise sit just before the
// "?" is removed as well. This lets a base URL and a fragment each contribute
// parameters without producing two "?" separators. The result is a valid,
// mergeable URL.
func ExampleURLJoin_query() {
	fmt.Println(urljoin.URLJoin("http://example.com", "foo?a=1", "?b=2"))
	// Output: http://example.com/foo?a=1&b=2
}
