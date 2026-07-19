// Package kebabcase converts strings to kebab-case, a stdlib-only Go port of
// the npm "kebab-case" package (which is itself commonly interchanged with
// lodash.kebabcase). Kebab-case is the lower-cased, hyphen-delimited word form
// popular in URL slugs, CSS class and custom-property names, HTML data
// attributes, CLI flags, and file names, where a single ASCII hyphen visually
// separates words like the skewers of a kebab. The package exposes a single
// entry point, KebabCase, so it can be dropped in wherever the Node original
// was used without carrying any third-party dependencies.
//
// A caller reaches for this package whenever a human-authored or programmatic
// identifier arrives in some other casing convention (camelCase, PascalCase,
// snake_case, "Title Case", or a mix of these) and a canonical kebab-case form
// is required. Typical uses are deriving a stable slug from a page title,
// normalizing component names, or turning a struct field name into a
// configuration key. Because the transformation is deterministic and depends
// only on the input runes, the same string always maps to the same result.
//
// The algorithm walks the input rune by rune and builds the output in one
// pass. Word boundaries are detected in two ways: an explicit separator (a
// space, an underscore, or an existing hyphen) is emitted as a hyphen, and an
// implicit boundary is introduced by writing a hyphen before any uppercase
// letter that immediately follows a lowercase letter or a digit. Every letter
// is then lower-cased as it is written. A second pass, trimAndCollapse,
// collapses any run of consecutive hyphens down to a single hyphen and strips
// hyphens from the two ends, so the result never begins, ends, or contains a
// doubled delimiter.
//
// The case rule is intentionally narrow, which gives it well-defined behavior
// on the awkward inputs. A run of adjacent capitals is treated as one word
// because no lowercase or digit precedes each capital, so "FOO" stays "foo"
// and an acronym-prefixed name such as "XMLHttpRequest" splits only where a
// lowercase precedes a capital, yielding "xmlhttp-request". A digit that
// precedes a capital does count as a boundary, so "foo123Bar" becomes
// "foo123-bar". Empty or whitespace-only input, and input consisting solely of
// separators such as "__foo__bar__" or "--foo--bar--", reduce cleanly to "" or
// to the trimmed "foo-bar" respectively.
//
// Parity with the Node original is close for the common ASCII identifiers that
// motivate the library: "fooBar Baz" becomes "foo-bar-baz" and an
// already-kebab string round-trips unchanged. Two differences are worth noting.
// Case folding here uses Go's unicode.ToLower and the boundary tests use
// unicode.IsUpper, IsLower, and IsDigit, so classification follows Unicode
// rather than the JavaScript regular expressions used upstream; and this port
// does not insert the extra boundaries that lodash.kebabcase derives from more
// elaborate word-splitting, so its splitting is limited to case changes and the
// space, underscore, and hyphen separators described above.
package kebabcase

import (
	"strings"
	"unicode"
)

// KebabCase converts s to kebab-case.
func KebabCase(s string) string {
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		switch {
		case !(unicode.IsLetter(r) || unicode.IsDigit(r)):
			// Any character that is neither a letter nor a digit acts as a word
			// separator, mirroring the upstream change-case strip regexp
			// /[^\p{L}\d]+/ rather than only recognizing space, underscore and
			// hyphen.
			b.WriteRune('-')
		case unicode.IsUpper(r):
			if i > 0 {
				prev := runes[i-1]
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					b.WriteRune('-')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return trimAndCollapse(b.String())
}

// trimAndCollapse collapses runs of hyphens into a single hyphen and removes any
// leading or trailing hyphens.
func trimAndCollapse(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if r == '-' {
			if prevDash {
				continue
			}
			prevDash = true
		} else {
			prevDash = false
		}
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), "-")
}
