// Package accepts performs HTTP content negotiation based on the Accept,
// Accept-Language, Accept-Charset and Accept-Encoding request headers. It is a
// port of the npm accepts package (the higher-level wrapper around
// negotiator that Express uses for req.accepts, req.acceptsLanguages,
// req.acceptsCharsets and req.acceptsEncodings) built with only the Go
// standard library.
//
// You reach for this package on the server side whenever a single endpoint can
// respond in more than one representation and you want the client's stated
// preferences to pick which one to send. Typical uses are choosing between
// JSON and HTML for the same resource, selecting a UI language, agreeing on a
// character set, or deciding whether to compress a response with gzip. Rather
// than parsing the raw header strings yourself, you hand the request headers to
// New and then ask, for each dimension, which of the representations you are
// able to produce ("offers") the client prefers.
//
// Like the Node original, this package is a thin layer: all of the header
// parsing, quality weighting, specificity scoring and tie-breaking lives in the
// sibling negotiator package, and accepts only adds the two conveniences that
// make it pleasant to call from a handler. The first is extension-style offers:
// an offer with no "/" in it is resolved through the mimetypes package, so
// Type("html", "json") works the same way Type("text/html",
// "application/json") does — and an extension nothing knows about is dropped
// rather than guessed at, exactly as mime.lookup returning false drops it
// upstream. The second is that the value you get back is the offer *you* passed
// in, with its original spelling, not the MIME type it was resolved to.
//
// Several edge cases mirror the Node semantics precisely. An offer whose
// effective quality is zero (for example against a spec written as
// application/json;q=0) is unacceptable and dropped, so the singular Type,
// Language, Charset and Encoding helpers return the empty string for it. When
// no offers are supplied the plural methods report the client's own preferences
// from the header instead, in preference order. When the Accept header is
// missing or empty the media-type negotiation short-circuits and the first
// offer is chosen unfiltered, which is what upstream's `if (!this.headers.accept)
// return types[0]` does. Encoding is special: the identity (no-transform)
// encoding is always acceptable unless explicitly disabled with identity;q=0 or
// a *;q=0 catch-all.
//
// Compared with the Node original the public surface is intentionally close:
// the plural methods correspond to accepts.types/languages/charsets/encodings
// and the singular ones to their first-result convenience forms. The deliberate
// differences are idiomatic Go ones. Offers are passed as variadic string
// arguments instead of an array; the singular methods return "" rather than
// false when nothing is acceptable; and the plural methods return the whole
// ranked list even when offers were supplied, where the Node function is
// overloaded to return only the best one. Taking the first element of the
// plural result is therefore equivalent to calling the Node function.
package accepts

import (
	"net/http"
	"strings"

	"github.com/malcolmston/express/mimetypes"
	"github.com/malcolmston/express/negotiator"
)

// Accepts negotiates acceptable response representations for a request based
// on its Accept* headers.
type Accepts struct {
	header http.Header
	neg    *negotiator.Negotiator
}

// New creates an Accepts from the given request headers, reading the Accept,
// Accept-Language, Accept-Charset and Accept-Encoding fields.
func New(header http.Header) *Accepts {
	if header == nil {
		header = http.Header{}
	}
	return &Accepts{header: header, neg: negotiator.New(header)}
}

// extToMime resolves an extension-style offer to a full MIME type. An offer
// that already contains a "/" is passed through untouched; anything else goes
// through the mimetypes table, and an unknown extension yields "" so the offer
// can be discarded rather than guessed at (upstream's mime.lookup returns false
// and validMime filters it out).
func extToMime(offer string) string {
	if strings.ContainsRune(offer, '/') {
		return offer
	}
	if t, ok := mimetypes.Lookup(offer); ok {
		return t
	}
	return ""
}

// Types returns the offers that are acceptable per the Accept header, ordered
// by preference. With no offers it returns the acceptable media types named in
// the header itself, in preference order.
func (a *Accepts) Types(offers ...string) []string {
	if len(offers) == 0 {
		return a.neg.MediaTypes()
	}

	// No Accept header (absent or empty): everything is acceptable and the
	// first offer wins, unfiltered.
	if a.header.Get("Accept") == "" {
		return append([]string(nil), offers...)
	}

	// Resolve each offer to a MIME type, keeping the positional mapping so the
	// caller's own spelling can be handed back.
	mimes := make([]string, len(offers))
	valid := make([]string, 0, len(offers))
	for i, offer := range offers {
		mimes[i] = extToMime(offer)
		if mimes[i] != "" {
			valid = append(valid, mimes[i])
		}
	}
	if len(valid) == 0 {
		// Every offer was an unknown extension. Upstream passes the resulting
		// empty array through, which negotiates to nothing — it must not be
		// confused with the no-arguments "list the header" form.
		return nil
	}

	accepted := a.neg.MediaTypes(valid...)
	out := make([]string, 0, len(accepted))
	for _, m := range accepted {
		for i, mime := range mimes {
			if mime == m {
				out = append(out, offers[i])
				break
			}
		}
	}
	return out
}

// Type returns the single best media type from offers, or "" if none are
// acceptable.
func (a *Accepts) Type(offers ...string) string {
	return first(a.Types(offers...))
}

// Languages returns the acceptable languages from offers in preference order,
// or the header's own languages when no offers are given.
func (a *Accepts) Languages(offers ...string) []string {
	return a.neg.Languages(offers...)
}

// Language returns the single best language from offers, or "".
func (a *Accepts) Language(offers ...string) string {
	return first(a.Languages(offers...))
}

// Charsets returns the acceptable charsets from offers in preference order, or
// the header's own charsets when no offers are given.
func (a *Accepts) Charsets(offers ...string) []string {
	return a.neg.Charsets(offers...)
}

// Charset returns the single best charset from offers, or "".
func (a *Accepts) Charset(offers ...string) string {
	return first(a.Charsets(offers...))
}

// Encodings returns the acceptable encodings from offers in preference order,
// or the header's own encodings when no offers are given. The "identity"
// encoding is always considered acceptable unless it is explicitly disallowed
// (identity;q=0 or *;q=0).
func (a *Accepts) Encodings(offers ...string) []string {
	return a.neg.Encodings(offers...)
}

// Encoding returns the single best encoding from offers, or "".
func (a *Accepts) Encoding(offers ...string) string {
	return first(a.Encodings(offers...))
}

func first(vs []string) string {
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}
