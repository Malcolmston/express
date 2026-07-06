package mimetypes_test

import (
	"fmt"

	"github.com/malcolmston/express/mimetypes"
)

// ExampleLookup resolves a filename to its MIME type. Lookup accepts a full
// path, a bare filename, a ".ext" or a bare extension, and returns the mapped
// type without any charset parameter. The second return value reports whether
// a mapping was found, letting callers distinguish a real answer from an
// unknown extension. Here a filename with a ".png" extension resolves to
// image/png. Passing an unknown extension would instead yield ("", false).
func ExampleLookup() {
	t, ok := mimetypes.Lookup("photo.png")
	fmt.Println(t, ok)
	// Output: image/png true
}

// ExampleLookup_unknown demonstrates the behavior for an extension that is not
// in the bundled table. Rather than guessing or returning a default type,
// Lookup reports failure through its boolean result. The returned string is
// empty and the boolean is false. This lets callers decide on their own
// fallback, such as application/octet-stream. The example prints both values
// to make the "unknown" signal explicit.
func ExampleLookup_unknown() {
	t, ok := mimetypes.Lookup("archive.unknownext")
	fmt.Printf("%q %v\n", t, ok)
	// Output: "" false
}

// ExampleContentType builds a full Content-Type header value from an
// extension. ContentType looks up the extension, then appends a
// "; charset=utf-8" parameter for textual types such as HTML. If the input
// already contained a "/" it would be treated as a MIME type directly. The
// boolean result reports whether a value could be produced. Here the "html"
// extension becomes a ready-to-send header string.
func ExampleContentType() {
	ct, ok := mimetypes.ContentType("html")
	fmt.Println(ct, ok)
	// Output: text/html; charset=utf-8 true
}

// ExampleExtension performs the reverse lookup, from a MIME type back to its
// canonical file extension. Any parameters on the type are ignored, so a
// charset suffix does not affect the result. Where several extensions share a
// type, the most common one is returned: image/jpeg maps to "jpg". The
// boolean is false for a type with no known extension. The extension is
// returned without a leading dot.
func ExampleExtension() {
	ext, ok := mimetypes.Extension("image/jpeg")
	fmt.Println(ext, ok)
	// Output: jpg true
}

// ExampleCharset reports the charset that applies to a MIME type. Every text/*
// type, along with JSON, XML, JavaScript and any "+json" or "+xml"
// structured-suffix type, is treated as UTF-8. Types with no associated
// charset, such as image/png, return ("", false). This is the same rule
// ContentType uses when deciding whether to append a charset parameter. Here
// application/json is reported as UTF-8.
func ExampleCharset() {
	cs, ok := mimetypes.Charset("application/json")
	fmt.Println(cs, ok)
	// Output: utf-8 true
}
