// Package urljoin joins URL path segments together, normalizing slashes. It is
// a standard-library-only Go port of the npm "url-join" package, the small
// utility countless Node projects use to assemble a URL from a base and one or
// more path fragments without ending up with doubled or missing slashes. The
// problem it solves is that naive string concatenation of URL parts produces
// artifacts like "http://host//api" or "api/v1users" depending on which parts
// carried trailing or leading slashes; url-join normalizes all of that away.
//
// URLJoin takes a variadic list of parts and returns a single joined URL.
// Segments are joined with exactly one "/", runs of duplicate slashes between
// segments are collapsed to a single slash, leading slashes are stripped from
// every part after the first, and trailing slashes are stripped from every part
// before the last. The last part keeps at most a single trailing slash, so an
// intentional trailing "/" survives while accidental doubles do not. Empty
// parts are skipped entirely, so passing "" between real segments does not
// introduce an empty path component.
//
// The protocol separator is treated specially so normalization does not damage
// it. The "://" that follows a scheme such as "http" or "https" is preserved
// rather than collapsed to a single slash, and a part that is a bare protocol
// like "http://" is merged with the following part before joining. The "file"
// scheme is handled distinctly because it canonically uses three slashes:
// "file:///" keeps its triple slash while other schemes are normalized to the
// usual "://" form.
//
// Query-string and fragment handling is likewise aware of URL structure. A
// slash that would otherwise sit immediately before a "?", "&", or "#"
// separator is removed so the query does not begin with a stray slash. When
// more than one part contributes a query string, the first "?" is kept as the
// query introducer and every subsequent "?" is rewritten to "&", so joining a
// base carrying "?a=1" with a fragment carrying "?b=2" yields a single
// well-formed "?a=1&b=2" query rather than two "?" separators.
//
// Parity with the Node original covers the slash collapsing, protocol
// preservation, file-scheme triple slash, and query-combining behavior that make
// url-join useful. The implementation uses the same regular-expression-driven
// approach as the reference library, and the API is reduced to a single
// idiomatic variadic function, URLJoin, that returns the empty string when called
// with no parts.
package urljoin

import (
	"regexp"
	"strings"
)

var (
	// A bare protocol such as "http://" or "file:///".
	reProtocolOnly = regexp.MustCompile(`^[^/:]+:/*$`)
	// The file protocol requires three slashes.
	reFileProtocol = regexp.MustCompile(`^file:///`)
	// Normalizes the protocol slashes of the first component.
	reProtocolSlashes = regexp.MustCompile(`^([^/:]+):/*`)
	// Leading and trailing slash runs.
	reLeadingSlashes  = regexp.MustCompile(`^[/]+`)
	reTrailingSlashes = regexp.MustCompile(`[/]+$`)
	// A slash immediately before a query/hash separator.
	reSlashBeforeParam = regexp.MustCompile(`/(\?|&|#[^!])`)
)

// URLJoin joins the given URL parts into a single URL, collapsing duplicate
// slashes while preserving the protocol separator and correctly combining
// query-string fragments.
func URLJoin(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}

	strArray := make([]string, len(parts))
	copy(strArray, parts)

	// If the first part is a bare protocol, merge it with the next part.
	if reProtocolOnly.MatchString(strArray[0]) && len(strArray) > 1 {
		first := strArray[0]
		strArray = strArray[1:]
		strArray[0] = first + strArray[0]
	}

	// Normalize the protocol slashes: file:// needs three, everything else two.
	if reFileProtocol.MatchString(strArray[0]) {
		strArray[0] = reProtocolSlashes.ReplaceAllString(strArray[0], "$1:///")
	} else {
		strArray[0] = reProtocolSlashes.ReplaceAllString(strArray[0], "$1://")
	}

	var resultArray []string
	for i := 0; i < len(strArray); i++ {
		component := strArray[i]
		if component == "" {
			continue
		}
		if i > 0 {
			// Strip leading slashes from every component but the first.
			component = reLeadingSlashes.ReplaceAllString(component, "")
		}
		if i < len(strArray)-1 {
			// Strip trailing slashes from every component but the last.
			component = reTrailingSlashes.ReplaceAllString(component, "")
		} else {
			// Collapse trailing slashes of the last component to a single one.
			component = reTrailingSlashes.ReplaceAllString(component, "/")
		}
		resultArray = append(resultArray, component)
	}

	str := strings.Join(resultArray, "/")

	// Remove a trailing slash that ended up before a query or hash separator.
	str = reSlashBeforeParam.ReplaceAllString(str, "$1")

	// Combine query strings: the first "?" stays, subsequent ones become "&".
	segs := strings.Split(str, "?")
	if len(segs) > 1 {
		str = segs[0] + "?" + strings.Join(segs[1:], "&")
	}

	return str
}
