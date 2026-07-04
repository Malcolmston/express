// Package urljoin joins URL path segments together, normalizing slashes.
//
// It is a faithful port of the npm package url-join. Segments are joined with a
// single "/", duplicate slashes are collapsed (while the "://" of a protocol is
// preserved), and query strings are combined using "?" and "&" separators.
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
