package express

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/malcolmston/express/httperrors"
)

// Accepts returns the best match among the offered types for the request's
// Accept header, or "" when none is acceptable. Offers may be extensions
// ("html", "json") or full media types ("text/html"). With no offers it
// returns the client's most-preferred type.
func (req *Request) Accepts(offers ...string) string {
	return negotiate(req.Get("Accept"), offers, mimeOf)
}

// AcceptsLanguages returns the best match among the offered languages for the
// Accept-Language header, or "". Matching honours language ranges: an offer of
// "en" satisfies an Accept-Language of "en-GB" (and vice versa), mirroring the
// negotiator prefix rules Express uses.
func (req *Request) AcceptsLanguages(offers ...string) string {
	return negotiateWith(req.Get("Accept-Language"), offers, languageMatches)
}

// AcceptsCharsets returns the best match among the offered charsets for the
// Accept-Charset header, or "".
func (req *Request) AcceptsCharsets(offers ...string) string {
	return negotiate(req.Get("Accept-Charset"), offers, func(s string) string { return s })
}

// AcceptsEncodings returns the best match among the offered encodings for the
// Accept-Encoding header, or "". As in Express, "identity" is always implicitly
// acceptable (unless explicitly rejected), and — unlike the other Accept
// families — an absent header does not make every offer acceptable: only
// identity is, so acceptsEncodings("gzip","identity") is "identity" with no
// Accept-Encoding present.
func (req *Request) AcceptsEncodings(offers ...string) string {
	return negotiateEncoding(req.Get("Accept-Encoding"), offers)
}

// languageMatches reports whether an Accept-Language entry matches an offered
// language, honouring prefix ranges ("en-GB" ~ "en") and the "*" wildcard.
func languageMatches(accept, offer string) bool {
	accept = strings.ToLower(accept)
	offer = strings.ToLower(offer)
	if accept == "*" || accept == offer {
		return true
	}
	aPrefix := accept
	if i := strings.IndexByte(accept, '-'); i >= 0 {
		aPrefix = accept[:i]
	}
	oPrefix := offer
	if i := strings.IndexByte(offer, '-'); i >= 0 {
		oPrefix = offer[:i]
	}
	return aPrefix == offer || accept == oPrefix
}

// negotiateWith is negotiate with a caller-supplied matcher and an identity
// resolver (the offer is compared to Accept entries directly). With no header it
// returns the first offer.
func negotiateWith(header string, offers []string, matches func(accept, offer string) bool) string {
	items := parseAccept(header)
	if len(offers) == 0 {
		for _, it := range items {
			if it.q > 0 {
				return it.value
			}
		}
		return ""
	}
	if header == "" {
		return offers[0]
	}
	best := ""
	bestQ := 0.0
	bestIdx := 1 << 30
	for oi, offer := range offers {
		for _, it := range items {
			if it.q <= 0 {
				continue
			}
			if matches(it.value, offer) && (it.q > bestQ || (it.q == bestQ && oi < bestIdx)) {
				best = offer
				bestQ = it.q
				bestIdx = oi
			}
		}
	}
	return best
}

// negotiateEncoding selects the best Accept-Encoding offer, treating identity as
// implicitly acceptable at the header's minimum quality unless a "*" or an
// explicit identity entry is present.
func negotiateEncoding(header string, offers []string) string {
	items := parseAccept(header)
	hasIdentity, hasStar := false, false
	minQ := 1.0
	for _, it := range items {
		switch it.value {
		case "identity":
			hasIdentity = true
		case "*":
			hasStar = true
		}
		if it.q < minQ {
			minQ = it.q
		}
	}
	qOf := func(enc string) (float64, bool) {
		enc = strings.ToLower(enc)
		best := -1.0
		for _, it := range items {
			if it.q <= 0 {
				continue
			}
			if it.value == enc || it.value == "*" {
				if it.q > best {
					best = it.q
				}
			}
		}
		if enc == "identity" && !hasIdentity && !hasStar && best < 0 {
			best = minQ // identity is implicitly acceptable
		}
		if best <= 0 {
			return 0, false
		}
		return best, true
	}
	bestOffer := ""
	bestQ := 0.0
	for _, offer := range offers {
		if q, ok := qOf(offer); ok && q > bestQ {
			bestOffer = offer
			bestQ = q
		}
	}
	return bestOffer
}

// acceptItem is one entry of an Accept-style header with its quality weight.
type acceptItem struct {
	value string
	q     float64
	order int
}

// parseAccept parses a comma-separated Accept-style header into weighted items,
// sorted by descending quality (stable on original order).
func parseAccept(header string) []acceptItem {
	var items []acceptItem
	for i, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		value := part
		q := 1.0
		if semi := strings.IndexByte(part, ';'); semi >= 0 {
			value = strings.TrimSpace(part[:semi])
			for _, param := range strings.Split(part[semi+1:], ";") {
				param = strings.TrimSpace(param)
				if strings.HasPrefix(param, "q=") {
					if v, err := strconv.ParseFloat(param[2:], 64); err == nil {
						q = v
					}
				}
			}
		}
		items = append(items, acceptItem{value: strings.ToLower(value), q: q, order: i})
	}
	sort.SliceStable(items, func(a, b int) bool { return items[a].q > items[b].q })
	return items
}

// negotiate picks the best offer against a header. resolve maps an offer to the
// media/value form compared against header entries (e.g. "html" -> "text/html").
func negotiate(header string, offers []string, resolve func(string) string) string {
	items := parseAccept(header)

	if len(offers) == 0 {
		for _, it := range items {
			if it.q > 0 {
				return it.value
			}
		}
		return ""
	}
	if header == "" {
		return offers[0] // no preference expressed
	}

	best := ""
	bestQ := 0.0
	bestOrder := 1 << 30
	for _, offer := range offers {
		resolved := strings.ToLower(resolve(offer))
		for _, it := range items {
			if it.q <= 0 {
				continue
			}
			if acceptMatches(it.value, resolved) && (it.q > bestQ || (it.q == bestQ && it.order < bestOrder)) {
				best = offer
				bestQ = it.q
				bestOrder = it.order
			}
		}
	}
	return best
}

// acceptMatches reports whether an Accept entry matches a resolved offer,
// honoring wildcards ("*/*", "text/*", "*").
func acceptMatches(accept, offer string) bool {
	if accept == "*" || accept == "*/*" {
		return true
	}
	if accept == offer {
		return true
	}
	// type/* wildcard.
	if strings.HasSuffix(accept, "/*") {
		prefix := strings.TrimSuffix(accept, "*")
		return strings.HasPrefix(offer, prefix)
	}
	return false
}

// mimeOf maps a short type name to a media type; full media types pass through.
func mimeOf(t string) string {
	if strings.Contains(t, "/") {
		return t
	}
	switch t {
	case "json":
		return "application/json"
	case "html":
		return "text/html"
	case "text", "txt":
		return "text/plain"
	case "xml":
		return "application/xml"
	case "js", "javascript":
		return "application/javascript"
	case "css":
		return "text/css"
	case "form":
		return "application/x-www-form-urlencoded"
	default:
		return t
	}
}

// Format performs content negotiation, invoking the handler for the best type
// the client accepts. Keys are extensions or media types; a "default" key (or
// the first entry) is used when nothing matches. It responds 406 when there is
// no match and no default.
func (res *Response) Format(handlers map[string]func()) {
	offers := make([]string, 0, len(handlers))
	for k := range handlers {
		if k != "default" {
			offers = append(offers, k)
		}
	}
	sort.Strings(offers)
	res.Vary("Accept")

	if best := res.req.Accepts(offers...); best != "" {
		if fn, ok := handlers[best]; ok {
			// Express sets the response Content-Type to the negotiated type
			// before running the branch, so a branch that just sends a string
			// inherits e.g. text/plain rather than the text/html default.
			res.Type(best)
			fn()
			return
		}
	}
	if fn, ok := handlers["default"]; ok {
		fn()
		return
	}
	// Nothing is acceptable and there is no default: Express hands a 406 error to
	// the middleware chain (next(NotAcceptableError)) rather than answering the
	// 406 itself, so an application error handler can shape the response.
	err := httperrors.New(http.StatusNotAcceptable, "Not Acceptable")
	if res.next != nil {
		res.next(err)
		return
	}
	res.Status(406).Send("Not Acceptable")
}

// Range represents a single byte range parsed from a Range header.
type Range struct {
	Start int64
	End   int64 // inclusive
}

// Ranges parses the request Range header against a resource of the given size,
// returning the satisfiable byte ranges. It returns (nil, false) when there is
// no Range header, and (nil, true) when the header is present but unsatisfiable.
func (req *Request) Ranges(size int64) ([]Range, bool) {
	h := req.Get("Range")
	if h == "" || !strings.HasPrefix(h, "bytes=") {
		return nil, false
	}
	spec := strings.TrimPrefix(h, "bytes=")
	var ranges []Range
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		dash := strings.IndexByte(part, '-')
		if dash < 0 {
			return nil, true
		}
		startStr, endStr := part[:dash], part[dash+1:]
		var start, end int64
		switch {
		case startStr == "": // suffix range: -N (last N bytes)
			n, err := strconv.ParseInt(endStr, 10, 64)
			if err != nil {
				return nil, true
			}
			if n > size {
				n = size
			}
			start, end = size-n, size-1
		case endStr == "": // start to end of resource
			s, err := strconv.ParseInt(startStr, 10, 64)
			if err != nil {
				return nil, true
			}
			start, end = s, size-1
		default:
			s, err1 := strconv.ParseInt(startStr, 10, 64)
			e, err2 := strconv.ParseInt(endStr, 10, 64)
			if err1 != nil || err2 != nil {
				return nil, true
			}
			start, end = s, e
			if end >= size {
				end = size - 1
			}
		}
		if start > end || start < 0 || start >= size {
			return nil, true // unsatisfiable
		}
		ranges = append(ranges, Range{Start: start, End: end})
	}
	if len(ranges) == 0 {
		return nil, true
	}
	return ranges, true
}
