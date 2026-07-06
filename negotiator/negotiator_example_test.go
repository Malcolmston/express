package negotiator_test

import (
	"fmt"
	"net/http"

	"github.com/malcolmston/express/negotiator"
)

// ExampleNegotiator_MediaType shows how to pick the single best response
// representation for a request. The client's Accept header lists two media
// types with explicit quality values, weighting JSON slightly above HTML. The
// server offers both forms and lets the negotiator choose. Because the higher
// q-value wins, application/json is selected even though text/html appears
// first in the header. MediaType returns the empty string when none of the
// offered types are acceptable.
func ExampleNegotiator_MediaType() {
	h := http.Header{}
	h.Set("Accept", "text/html;q=0.8, application/json;q=0.9")

	n := negotiator.New(h)
	fmt.Println(n.MediaType("text/html", "application/json"))
	// Output: application/json
}

// ExampleNegotiator_MediaTypes shows the plural form, which returns every
// acceptable media type in descending preference order rather than just the
// best one. The same weighted Accept header is used, so the results are
// ordered by quality with the higher-weighted JSON first. This is useful when
// the caller wants to attempt each representation in turn. Ties on quality
// preserve the order in which the available types were passed. The returned
// slice is empty when nothing matches.
func ExampleNegotiator_MediaTypes() {
	h := http.Header{}
	h.Set("Accept", "text/html;q=0.8, application/json;q=0.9")

	n := negotiator.New(h)
	fmt.Println(n.MediaTypes("text/html", "application/json"))
	// Output: [application/json text/html]
}

// ExampleNegotiator_Language demonstrates language negotiation over the
// Accept-Language header. The client prefers French over English, expressed
// through the quality values attached to each tag. The server can serve either
// language and asks the negotiator which to use. Since French carries the
// higher q-value it is chosen. Language matching is also prefix aware, so a
// header of "en" would match an available "en-US".
func ExampleNegotiator_Language() {
	h := http.Header{}
	h.Set("Accept-Language", "en;q=0.5, fr;q=0.9")

	n := negotiator.New(h)
	fmt.Println(n.Language("en", "fr"))
	// Output: fr
}

// ExampleNegotiator_Charset demonstrates charset negotiation over the
// Accept-Charset header. The client accepts UTF-8 at full quality and
// ISO-8859-1 at a reduced quality. The server offers both charsets in the
// less-preferred order to show that ordering is driven by quality, not by the
// argument order. UTF-8 is therefore returned as the best match. Charset
// returns the empty string when no offered charset is acceptable.
func ExampleNegotiator_Charset() {
	h := http.Header{}
	h.Set("Accept-Charset", "utf-8, iso-8859-1;q=0.5")

	n := negotiator.New(h)
	fmt.Println(n.Charset("iso-8859-1", "utf-8"))
	// Output: utf-8
}

// ExampleNegotiator_Encodings demonstrates content-encoding negotiation over
// the Accept-Encoding header. The client advertises gzip and brotli with
// brotli weighted higher. The server supports both compression schemes and
// asks for all acceptable encodings in order. The result lists brotli first
// because of its higher quality value. Note that the "identity" (uncompressed)
// encoding is always considered acceptable unless it is explicitly disabled
// with identity;q=0.
func ExampleNegotiator_Encodings() {
	h := http.Header{}
	h.Set("Accept-Encoding", "gzip;q=0.8, br;q=0.9")

	n := negotiator.New(h)
	fmt.Println(n.Encodings("gzip", "br"))
	// Output: [br gzip]
}
