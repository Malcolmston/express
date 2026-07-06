package contentrange_test

import (
	"fmt"

	"github.com/malcolmston/express/contentrange"
)

// ExampleFormat builds a Content-Range header value from its numeric parts. The
// header has the form "<unit> <range>/<size>", so the start, end, and total size
// render as "bytes 0-499/1234". An empty unit defaults to "bytes". Servers send
// this header on 206 Partial Content responses to describe which slice of a
// representation the body carries.
func ExampleFormat() {
	fmt.Println(contentrange.Format("bytes", 0, 499, 1234))
	// Output: bytes 0-499/1234
}

// ExampleParse decodes a Content-Range header value into a ContentRange struct.
// It splits the unit off at the first space and the range from the size at the
// '/'. Concrete numeric values set the HasRange and HasSize flags to true; a "*"
// clears the corresponding flag. Here a fully concrete header yields the start,
// end, and size fields directly.
func ExampleParse() {
	cr, err := contentrange.Parse("bytes 0-499/1234")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cr.Start, cr.End, cr.Size)
	// Output: 0 499 1234
}

// ExampleContentRange_String renders a ContentRange back into a header value,
// the inverse of Parse. Because the numeric fields double as sentinels, the
// HasRange and HasSize flags decide whether concrete values or "*" are emitted.
// Here HasSize is false, so the size is rendered as "*" to signal an unknown
// total length while the concrete range is preserved.
func ExampleContentRange_String() {
	cr := contentrange.ContentRange{
		Unit:     "bytes",
		Start:    0,
		End:      499,
		HasRange: true,
		HasSize:  false,
	}
	fmt.Println(cr.String())
	// Output: bytes 0-499/*
}
