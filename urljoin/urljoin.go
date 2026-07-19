// Package urljoin joins URL path segments together, normalizing slashes. It is
// a standard-library-only Go port of the npm "url-join" package, the small
// utility countless Node projects use to assemble a URL from a base and one or
// more path fragments without ending up with doubled or missing slashes. The
// problem it solves is that naive string concatenation of URL parts produces
// artifacts like "http://host//api" or "api/v1users" depending on which parts
// carried trailing or leading slashes; url-join normalizes all of that away.
//
// URLJoin takes a variadic list of parts and returns a single joined URL.
// Empty parts are filtered out entirely, segments are joined with exactly one
// "/", runs of duplicate slashes between segments are collapsed to a single
// slash, leading slashes are stripped from every part after the first, and
// trailing slashes are stripped from every part before the last. The last part
// keeps at most a single trailing slash, so an intentional trailing "/"
// survives while accidental doubles do not.
//
// The protocol separator is treated specially so normalization does not damage
// it. A part that is a bare protocol such as "http://" or "http:" is merged
// with the following part before joining, and a leading lone "/" is likewise
// merged with the next part. The "file" scheme is handled distinctly because it
// canonically uses three slashes: a "file:///" prefix keeps its triple slash
// while other schemes are normalized to the usual "://" form. A leading IPv6
// host in brackets (for example "[2601:195:c381:3560::f42a]") is left untouched
// so its inner colons are not mistaken for a protocol separator.
//
// Query-string and fragment handling is likewise aware of URL structure. When a
// segment ends with "?" or "#", the following segment is appended without an
// intervening slash, and a slash that would otherwise sit immediately before a
// "?", "&", or "#" separator is removed. Everything after the first "#" is
// treated as an opaque fragment and left alone. Within the portion before the
// fragment, the run of query parameters introduced by any mix of "?" and "&" is
// re-stitched into a single well-formed query: the first "?" introduces the
// query and every subsequent parameter is joined with "&", while empty
// parameters (a bare "?" or "&") are dropped.
//
// The implementation follows the same approach as the reference library
// (url-join 5.0.0), and the API is reduced to a single idiomatic variadic
// function, URLJoin, that returns the empty string when called with no parts.
package urljoin

import (
	"regexp"
	"strings"
)

var (
	// A bare protocol such as "http://", "http:" or "file:///".
	reProtocolOnly = regexp.MustCompile(`^[^/:]+:/*$`)
	// The file protocol requires three slashes.
	reFileProtocol = regexp.MustCompile(`^file:///`)
	// A leading IPv6 host in brackets, e.g. "[2601:195:c381:3560::f42a]".
	reIPv6Host = regexp.MustCompile(`^\[.*:.*\]`)
	// Normalizes the protocol slashes of the first component.
	reProtocolSlashes = regexp.MustCompile(`^([^/:]+):/*`)
	// Leading and trailing slash runs.
	reLeadingSlashes  = regexp.MustCompile(`^[/]+`)
	reTrailingSlashes = regexp.MustCompile(`[/]+$`)
	// A slash immediately before a query/hash separator.
	reSlashBeforeParam = regexp.MustCompile(`/(\?|&|#[^!])`)
	// A run of query separators ("?" and/or "&").
	reQuerySeparators = regexp.MustCompile(`[?&]+`)
)

// URLJoin joins the given URL parts into a single URL, collapsing duplicate
// slashes while preserving the protocol separator and correctly combining
// query-string fragments.
func URLJoin(parts ...string) string {
	// Filter out any empty string values.
	strArray := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			strArray = append(strArray, p)
		}
	}
	if len(strArray) == 0 {
		return ""
	}

	// If the first part is a bare protocol, merge it with the next part.
	if reProtocolOnly.MatchString(strArray[0]) && len(strArray) > 1 {
		strArray[1] = strArray[0] + strArray[1]
		strArray = strArray[1:]
	}

	// If the first part is a lone leading slash, merge it with the next part.
	if strArray[0] == "/" && len(strArray) > 1 {
		strArray[1] = strArray[0] + strArray[1]
		strArray = strArray[1:]
	}

	// There must be two or three slashes in the file protocol, two slashes in
	// anything else. A leading IPv6 host is left alone so its colons are not
	// treated as a protocol separator.
	if reFileProtocol.MatchString(strArray[0]) {
		strArray[0] = reProtocolSlashes.ReplaceAllString(strArray[0], "$1:///")
	} else if !reIPv6Host.MatchString(strArray[0]) {
		strArray[0] = reProtocolSlashes.ReplaceAllString(strArray[0], "$1://")
	}

	resultArray := make([]string, 0, len(strArray))
	for i := 0; i < len(strArray); i++ {
		component := strArray[i]
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
		if component == "" {
			continue
		}
		resultArray = append(resultArray, component)
	}

	// Join with a single slash, except that a component following one that ends
	// with "?" or "#" is appended directly so the query/fragment stays intact.
	var b strings.Builder
	for i, part := range resultArray {
		if i == 0 {
			b.WriteString(part)
			continue
		}
		prev := resultArray[i-1]
		if strings.HasSuffix(prev, "?") || strings.HasSuffix(prev, "#") {
			b.WriteString(part)
			continue
		}
		b.WriteByte('/')
		b.WriteString(part)
	}
	str := b.String()

	// Remove a trailing slash that ended up before a query or hash separator.
	str = reSlashBeforeParam.ReplaceAllString(str, "$1")

	// Split off the fragment: everything after the first "#" is opaque and is
	// preserved as-is. (This mirrors the reference library, which keeps only the
	// segment immediately following the first "#".)
	hashSegs := strings.Split(str, "#")
	beforeHash := hashSegs[0]
	afterHash := ""
	if len(hashSegs) > 1 {
		afterHash = hashSegs[1]
	}

	// Re-stitch the query parameters: the first "?" introduces the query and
	// every subsequent parameter is joined with "&". Empty parameters (from a
	// bare "?" or "&") are dropped.
	var queryParts []string
	for _, p := range reQuerySeparators.Split(beforeHash, -1) {
		if p != "" {
			queryParts = append(queryParts, p)
		}
	}

	var out strings.Builder
	if len(queryParts) > 0 {
		out.WriteString(queryParts[0])
		if len(queryParts) > 1 {
			out.WriteByte('?')
			out.WriteString(strings.Join(queryParts[1:], "&"))
		}
	}
	if afterHash != "" {
		out.WriteByte('#')
		out.WriteString(afterHash)
	}

	return out.String()
}
