package htmlentities_test

import (
	"fmt"

	"github.com/malcolmston/express/htmlentities"
)

// ExampleEncode shows the default "specialChars" mode, which escapes only the
// five characters that are significant in HTML: & < > " and '. This is the mode
// you want when inserting untrusted text into an HTML document, since it
// neutralizes markup and attribute-breaking characters. Non-special characters,
// including accented and other non-ASCII runes, are passed through unchanged.
// The result below is safe to drop into an HTML page. Calling Encode with no
// options selects this default behavior.
func ExampleEncode() {
	out := htmlentities.Encode(`<a href="x">Tom & Jerry's</a>`)
	fmt.Println(out)
	// Output: &lt;a href=&quot;x&quot;&gt;Tom &amp; Jerry&apos;s&lt;/a&gt;
}

// ExampleEncode_nonAscii demonstrates the stricter "nonAscii" mode, selected
// through EncodeOptions. In addition to escaping the five special characters,
// this mode rewrites every rune above code point 127 as a decimal numeric
// entity. That makes the output pure ASCII, which is useful for transports that
// are not guaranteed to be UTF-8 clean. Here the accented "é" becomes &#233;
// and the copyright sign becomes &#169;. The special characters would still be
// escaped in this mode if the input contained any.
func ExampleEncode_nonAscii() {
	out := htmlentities.Encode("café ©", htmlentities.EncodeOptions{Mode: "nonAscii"})
	fmt.Println(out)
	// Output: caf&#233; &#169;
}

// ExampleDecode converts HTML entities back into their literal characters. It
// resolves named entities from a built-in table as well as decimal (&#233;) and
// hexadecimal (&#xe9;) numeric references. Entities it does not recognize, such
// as an unknown name or a bare ampersand, are left exactly as they appear. This
// example decodes a mix of the five special named entities back into markup
// characters. Decoding is the inverse of specialChars-mode Encode for those
// characters, so escaped text round-trips cleanly.
func ExampleDecode() {
	out := htmlentities.Decode(`&lt;a href=&quot;x&quot;&gt;Tom &amp; Jerry&apos;s`)
	fmt.Println(out)
	// Output: <a href="x">Tom & Jerry's
}
