// Package accepts performs HTTP content negotiation based on the Accept,
// Accept-Language, Accept-Charset and Accept-Encoding request headers. It is a
// port of the npm accepts package using only the Go standard library.
package accepts

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Accepts negotiates acceptable response representations for a request based
// on its Accept* headers.
type Accepts struct {
	accept   string
	language string
	charset  string
	encoding string
}

// New creates an Accepts from the given request headers, reading the Accept,
// Accept-Language, Accept-Charset and Accept-Encoding fields.
func New(header http.Header) *Accepts {
	return &Accepts{
		accept:   header.Get("Accept"),
		language: header.Get("Accept-Language"),
		charset:  header.Get("Accept-Charset"),
		encoding: header.Get("Accept-Encoding"),
	}
}

// spec is a single parsed entry from an Accept* header.
type spec struct {
	value string  // full value, e.g. "text/html" or "en-US" or "gzip"
	main  string  // primary part (type for media, primary tag for lang)
	sub   string  // secondary part (subtype for media, sublang for lang)
	q     float64 // quality value
	order int     // original order, used as a tiebreaker
}

// mimeShorthand maps extension-style offer names to full MIME types.
var mimeShorthand = map[string]string{
	"json":       "application/json",
	"html":       "text/html",
	"htm":        "text/html",
	"text":       "text/plain",
	"txt":        "text/plain",
	"xml":        "application/xml",
	"css":        "text/css",
	"js":         "application/javascript",
	"urlencoded": "application/x-www-form-urlencoded",
	"multipart":  "multipart/form-data",
	"png":        "image/png",
	"jpg":        "image/jpeg",
	"jpeg":       "image/jpeg",
	"gif":        "image/gif",
}

// normalizeOffer maps a media-type offer (possibly a shorthand) to a full
// MIME type.
func normalizeOffer(offer string) string {
	o := strings.ToLower(strings.TrimSpace(offer))
	if full, ok := mimeShorthand[o]; ok {
		return full
	}
	if !strings.ContainsRune(o, '/') {
		if full, ok := mimeShorthand[o]; ok {
			return full
		}
		return "application/" + o
	}
	return o
}

// parseHeader parses an Accept* header into specs. splitFn splits a value into
// main/sub parts (sub may be empty).
func parseHeader(header string, splitFn func(string) (string, string)) []spec {
	var specs []spec
	parts := strings.Split(header, ",")
	order := 0
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		segs := strings.Split(p, ";")
		value := strings.TrimSpace(segs[0])
		if value == "" {
			continue
		}
		q := 1.0
		for _, param := range segs[1:] {
			param = strings.TrimSpace(param)
			if strings.HasPrefix(param, "q=") || strings.HasPrefix(param, "Q=") {
				if f, err := strconv.ParseFloat(strings.TrimSpace(param[2:]), 64); err == nil {
					q = f
				}
			}
		}
		main, sub := splitFn(value)
		specs = append(specs, spec{
			value: value,
			main:  main,
			sub:   sub,
			q:     q,
			order: order,
		})
		order++
	}
	return specs
}

// splitMedia splits "type/subtype" into its parts.
func splitMedia(v string) (string, string) {
	v = strings.ToLower(v)
	if i := strings.IndexByte(v, '/'); i >= 0 {
		return v[:i], v[i+1:]
	}
	return v, ""
}

// splitLang splits "primary-sub" into its parts.
func splitLang(v string) (string, string) {
	v = strings.ToLower(v)
	if i := strings.IndexByte(v, '-'); i >= 0 {
		return v[:i], v[i+1:]
	}
	return v, ""
}

// splitPlain treats the whole value as the main part.
func splitPlain(v string) (string, string) {
	return strings.ToLower(v), ""
}

// mediaQuality returns the q-value for a concrete media type given the parsed
// Accept specs, plus a specificity score, and whether it is acceptable.
func mediaQuality(typ, sub string, specs []spec) (q float64, spec2 int, ok bool) {
	best := -1.0
	bestSpec := -1
	for _, s := range specs {
		var score int
		switch {
		case s.main == "*" && s.sub == "*":
			score = 0
		case s.main == typ && s.sub == "*":
			score = 1
		case s.main == typ && s.sub == sub:
			score = 2
		default:
			continue
		}
		if score > bestSpec || (score == bestSpec && s.q > best) {
			best = s.q
			bestSpec = score
		}
	}
	if bestSpec < 0 {
		return 0, 0, false
	}
	return best, bestSpec, best > 0
}

// langQuality returns the q-value for a concrete language tag.
func langQuality(primary, sub string, specs []spec) (q float64, score int, ok bool) {
	best := -1.0
	bestScore := -1
	for _, s := range specs {
		var sc int
		switch {
		case s.main == "*":
			sc = 0
		case s.main == primary && s.sub == sub:
			sc = 2
		case s.main == primary && s.sub == "":
			sc = 1
		default:
			continue
		}
		if sc > bestScore || (sc == bestScore && s.q > best) {
			best = s.q
			bestScore = sc
		}
	}
	if bestScore < 0 {
		return 0, 0, false
	}
	return best, bestScore, best > 0
}

// plainQuality returns the q-value for a concrete charset/encoding token.
func plainQuality(value string, specs []spec) (q float64, exact bool, ok bool) {
	v := strings.ToLower(value)
	best := -1.0
	bestScore := -1
	for _, s := range specs {
		var sc int
		switch {
		case s.main == "*":
			sc = 0
		case s.main == v:
			sc = 1
		default:
			continue
		}
		if sc > bestScore || (sc == bestScore && s.q > best) {
			best = s.q
			bestScore = sc
		}
	}
	if bestScore < 0 {
		return 0, false, false
	}
	return best, bestScore == 1, best > 0
}

// scored pairs an offer with its negotiated quality for stable sorting.
type scored struct {
	offer string
	q     float64
	spec  int
	order int
}

// sortScored orders offers by descending quality, then descending
// specificity, then original order.
func sortScored(items []scored) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].q != items[j].q {
			return items[i].q > items[j].q
		}
		if items[i].spec != items[j].spec {
			return items[i].spec > items[j].spec
		}
		return items[i].order < items[j].order
	})
}

// Types returns the subset of offers that are acceptable per the Accept
// header, ordered by preference. With no offers it returns all acceptable
// media types from the header in preference order.
func (a *Accepts) Types(offers ...string) []string {
	specs := parseHeader(a.accept, splitMedia)

	if len(offers) == 0 {
		// Return the header's media types in preference order (excluding
		// q=0 and wildcards resolved as-is).
		var items []scored
		for _, s := range specs {
			if s.q <= 0 {
				continue
			}
			items = append(items, scored{offer: s.value, q: s.q, order: s.order})
		}
		sortScored(items)
		out := make([]string, 0, len(items))
		for _, it := range items {
			out = append(out, it.offer)
		}
		return out
	}

	// No Accept header means everything is acceptable.
	if strings.TrimSpace(a.accept) == "" {
		return append([]string(nil), offers...)
	}

	var items []scored
	for i, offer := range offers {
		full := normalizeOffer(offer)
		typ, sub := splitMedia(full)
		q, sc, ok := mediaQuality(typ, sub, specs)
		if !ok {
			continue
		}
		items = append(items, scored{offer: offer, q: q, spec: sc, order: i})
	}
	sortScored(items)
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.offer)
	}
	return out
}

// Type returns the single best media type from offers, or "" if none are
// acceptable.
func (a *Accepts) Type(offers ...string) string {
	res := a.Types(offers...)
	if len(res) == 0 {
		return ""
	}
	return res[0]
}

// Languages returns the acceptable languages from offers in preference order,
// or all header languages when no offers are given.
func (a *Accepts) Languages(offers ...string) []string {
	specs := parseHeader(a.language, splitLang)

	if len(offers) == 0 {
		var items []scored
		for _, s := range specs {
			if s.q <= 0 {
				continue
			}
			items = append(items, scored{offer: s.value, q: s.q, order: s.order})
		}
		sortScored(items)
		out := make([]string, 0, len(items))
		for _, it := range items {
			out = append(out, it.offer)
		}
		return out
	}

	if strings.TrimSpace(a.language) == "" {
		return append([]string(nil), offers...)
	}

	var items []scored
	for i, offer := range offers {
		primary, sub := splitLang(offer)
		q, sc, ok := langQuality(primary, sub, specs)
		if !ok {
			continue
		}
		items = append(items, scored{offer: offer, q: q, spec: sc, order: i})
	}
	sortScored(items)
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.offer)
	}
	return out
}

// Language returns the single best language from offers, or "".
func (a *Accepts) Language(offers ...string) string {
	res := a.Languages(offers...)
	if len(res) == 0 {
		return ""
	}
	return res[0]
}

// Charsets returns the acceptable charsets from offers in preference order, or
// all header charsets when no offers are given.
func (a *Accepts) Charsets(offers ...string) []string {
	specs := parseHeader(a.charset, splitPlain)

	if len(offers) == 0 {
		var items []scored
		for _, s := range specs {
			if s.q <= 0 {
				continue
			}
			items = append(items, scored{offer: s.value, q: s.q, order: s.order})
		}
		sortScored(items)
		out := make([]string, 0, len(items))
		for _, it := range items {
			out = append(out, it.offer)
		}
		return out
	}

	if strings.TrimSpace(a.charset) == "" {
		return append([]string(nil), offers...)
	}

	var items []scored
	for i, offer := range offers {
		q, exact, ok := plainQuality(offer, specs)
		if !ok {
			continue
		}
		sc := 0
		if exact {
			sc = 1
		}
		items = append(items, scored{offer: offer, q: q, spec: sc, order: i})
	}
	sortScored(items)
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.offer)
	}
	return out
}

// Charset returns the single best charset from offers, or "".
func (a *Accepts) Charset(offers ...string) string {
	res := a.Charsets(offers...)
	if len(res) == 0 {
		return ""
	}
	return res[0]
}

// Encodings returns the acceptable encodings from offers in preference order,
// or all header encodings when no offers are given. The "identity" encoding is
// always considered acceptable unless it is explicitly disallowed (q=0).
func (a *Accepts) Encodings(offers ...string) []string {
	specs := parseHeader(a.encoding, splitPlain)

	// Determine whether identity is allowed and its quality.
	identityQ := 1.0
	identitySet := false
	starQ := -1.0
	starSet := false
	for _, s := range specs {
		if s.main == "identity" {
			identityQ = s.q
			identitySet = true
		}
		if s.main == "*" {
			starQ = s.q
			starSet = true
		}
	}
	if !identitySet && starSet {
		identityQ = starQ
		identitySet = true
	}
	if !identitySet {
		identityQ = 1.0
	}

	if len(offers) == 0 {
		var items []scored
		seen := map[string]bool{}
		for _, s := range specs {
			if s.q <= 0 || s.main == "*" {
				continue
			}
			items = append(items, scored{offer: s.value, q: s.q, order: s.order})
			seen[strings.ToLower(s.value)] = true
		}
		if identityQ > 0 && !seen["identity"] {
			items = append(items, scored{offer: "identity", q: identityQ, order: len(specs)})
		}
		sortScored(items)
		out := make([]string, 0, len(items))
		for _, it := range items {
			out = append(out, it.offer)
		}
		return out
	}

	// No Accept-Encoding header: only identity is implied acceptable.
	if strings.TrimSpace(a.encoding) == "" {
		for _, offer := range offers {
			if strings.EqualFold(offer, "identity") {
				return []string{offer}
			}
		}
		// Per spec, absence means anything is acceptable; return offers as-is
		// with identity preferred already handled above.
		return append([]string(nil), offers...)
	}

	var items []scored
	for i, offer := range offers {
		if strings.EqualFold(offer, "identity") {
			if identityQ > 0 {
				items = append(items, scored{offer: offer, q: identityQ, spec: 1, order: i})
			}
			continue
		}
		q, exact, ok := plainQuality(offer, specs)
		if !ok || q <= 0 {
			continue
		}
		sc := 0
		if exact {
			sc = 1
		}
		items = append(items, scored{offer: offer, q: q, spec: sc, order: i})
	}
	sortScored(items)
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.offer)
	}
	return out
}

// Encoding returns the single best encoding from offers, or "".
func (a *Accepts) Encoding(offers ...string) string {
	res := a.Encodings(offers...)
	if len(res) == 0 {
		return ""
	}
	return res[0]
}
