package querystringlite_test

import (
	"fmt"

	"github.com/malcolmston/express/querystringlite"
)

// ExampleParse treats the query string as a flat multimap, so a repeated key
// keeps all of its values. Here "a" appears twice and maps to the ordered list
// of both values, while "b" maps to a single-element list. Unlike the qs
// package, "a[b]" would be a literal key rather than a request to build a nested
// object. The result is printed with fmt, which sorts the map keys for a stable
// rendering. This matches the behavior of Node's built-in querystring.parse.
func ExampleParse() {
	fmt.Println(querystringlite.Parse("a=1&b=2&a=3"))
	// Output: map[a:[1 3] b:[2]]
}

// ExampleStringify is the inverse of Parse and emits keys in sorted order for
// deterministic output. A multi-valued key expands to repeated "key=value"
// pairs, so the two values of "a" become two separate pairs. Both keys and
// values are percent-encoded via Escape, and a key with no values would be
// omitted entirely. The sorted ordering makes the output safe to compare in
// tests. Here the map serializes back to a canonical query string.
func ExampleStringify() {
	v := map[string][]string{"a": {"1", "3"}, "b": {"2"}}
	fmt.Println(querystringlite.Stringify(v))
	// Output: a=1&a=3&b=2
}

// ExampleEscape percent-encodes a string the way Node's querystring.escape
// does. Characters outside the unreserved set are replaced with their "%XX"
// hexadecimal form, and importantly a space becomes "%20" rather than "+".
// Here the space and the ampersand are both encoded while the letters pass
// through unchanged. This is the same encoding Stringify applies to keys and
// values. The uppercase hex digits match Node's output exactly.
func ExampleEscape() {
	fmt.Println(querystringlite.Escape("a b&c"))
	// Output: a%20b%26c
}
