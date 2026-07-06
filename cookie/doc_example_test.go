package cookie_test

import (
	"fmt"

	"github.com/malcolmston/express/cookie"
)

// ExampleSerialize builds a Set-Cookie header value from a name, a value, and a
// set of attributes. The value is percent-encoded using encodeURIComponent
// conventions, so the space and '+' in "a b+c" become "%20" and "%2B". The
// Options fields add Path, Max-Age, and HttpOnly attributes in the canonical
// order. This is the header a server sends to store a cookie in the client.
func ExampleSerialize() {
	s, err := cookie.Serialize("sid", "a b+c", &cookie.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(s)
	// Output: sid=a%20b%2Bc; Path=/; Max-Age=3600; HttpOnly
}

// ExampleParse turns a raw Cookie request header into a map of names to values.
// Each value is URL-decoded, reversing the encoding that Serialize applies, so
// "a%20b%2Bc" decodes back to "a b+c". When a name appears more than once the
// first occurrence wins. Here two cookies are parsed out of a single header
// string.
func ExampleParse() {
	m := cookie.Parse("sid=a%20b%2Bc; theme=dark")
	fmt.Printf("%q %q\n", m["sid"], m["theme"])
	// Output: "a b+c" "dark"
}
