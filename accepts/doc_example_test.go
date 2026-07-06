package accepts_test

import (
	"fmt"
	"net/http"

	"github.com/malcolmston/express/accepts"
)

// ExampleAccepts_Type shows server-side content negotiation for a response
// body. A request carries an Accept header listing the media types the client
// will take, each with an optional quality weight. New wraps the request
// headers and Type picks the single best match from the representations the
// server can actually produce. Here the client slightly prefers HTML over JSON,
// so Type returns "html" even though "json" was offered first.
func ExampleAccepts_Type() {
	h := http.Header{}
	h.Set("Accept", "text/html, application/json;q=0.9")

	a := accepts.New(h)
	fmt.Println(a.Type("json", "html"))
	// Output: html
}

// ExampleAccepts_Types returns every acceptable offer in preference order
// rather than just the top one. The offers may be short extension-style names
// such as "json" and "html", which are expanded to full MIME types internally.
// Results are ordered by descending quality and then by specificity, with ties
// falling back to the order the offers were passed. This is useful when the
// caller wants to try representations in turn rather than commit to one.
func ExampleAccepts_Types() {
	h := http.Header{}
	h.Set("Accept", "text/html, application/json;q=0.9")

	a := accepts.New(h)
	fmt.Println(a.Types("json", "html"))
	// Output: [html json]
}

// ExampleAccepts_Language negotiates a UI language from the Accept-Language
// header. Each language tag may carry a quality weight, and Language returns the
// offered tag with the highest weight. Here "en" outranks "fr" because it was
// given q=0.9 versus q=0.8, so the English offer is chosen. Absent or malformed
// weights default to 1.0, matching HTTP content negotiation.
func ExampleAccepts_Language() {
	h := http.Header{}
	h.Set("Accept-Language", "en-US, en;q=0.9, fr;q=0.8")

	a := accepts.New(h)
	fmt.Println(a.Language("fr", "en"))
	// Output: en
}

// ExampleAccepts_Encoding chooses a response compression scheme. The
// Accept-Encoding header lists the encodings a client understands, and Encoding
// returns the best offered match. The identity (uncompressed) encoding is always
// implicitly acceptable unless explicitly disabled, but here gzip is preferred
// over the lower-weighted deflate. This is the negotiation Express performs
// before compressing a response.
func ExampleAccepts_Encoding() {
	h := http.Header{}
	h.Set("Accept-Encoding", "gzip, deflate;q=0.5")

	a := accepts.New(h)
	fmt.Println(a.Encoding("gzip", "deflate"))
	// Output: gzip
}
