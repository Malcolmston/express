// Package negotiator is an HTTP content negotiation helper, a port of the
// npm "negotiator" package. It inspects request headers such as Accept,
// Accept-Language, Accept-Charset and Accept-Encoding and reports which of a
// set of available representations best match the client's preferences.
package negotiator

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Negotiator performs content negotiation against a set of request headers.
type Negotiator struct {
	header http.Header
}

// New constructs a Negotiator from the given request headers.
func New(header http.Header) *Negotiator {
	if header == nil {
		header = http.Header{}
	}
	return &Negotiator{header: header}
}

// spec represents a single parsed entry from an Accept-style header.
type spec struct {
	value string
	// for media types
	typ    string
	subtyp string
	params map[string]string
	q      float64
	order  int // original position, used for stable ordering
	// index of the matched available value, used when negotiating
	index int
}

// ---------------------------------------------------------------------------
// Media types
// ---------------------------------------------------------------------------

// MediaTypes returns the media types that are acceptable given the request's
// Accept header, restricted to the provided available media types. When called
// with no arguments it returns every acceptable media type in the header sorted
// by descending quality (stable by original order).
func (n *Negotiator) MediaTypes(available ...string) []string {
	accepts := parseAccept(n.header.Get("Accept"))
	if len(available) == 0 {
		var full []spec
		for _, a := range accepts {
			if a.q > 0 {
				full = append(full, a)
			}
		}
		sortSpecs(full)
		out := make([]string, 0, len(full))
		for _, a := range full {
			out = append(out, a.value)
		}
		return out
	}
	var matched []spec
	for i, av := range available {
		if best, ok := bestMediaMatch(accepts, av); ok && best.q > 0 {
			best.index = i
			best.value = av
			matched = append(matched, best)
		}
	}
	sortSpecsByMatch(matched)
	out := make([]string, 0, len(matched))
	for _, m := range matched {
		out = append(out, m.value)
	}
	return out
}

// MediaType returns the single best acceptable media type, or "" if none of the
// available media types are acceptable.
func (n *Negotiator) MediaType(available ...string) string {
	res := n.MediaTypes(available...)
	if len(res) == 0 {
		return ""
	}
	return res[0]
}

func parseAccept(header string) []spec {
	var specs []spec
	if strings.TrimSpace(header) == "" {
		header = "*/*"
	}
	for i, part := range splitComma(header) {
		typ, subtyp, params, q, ok := parseMediaRange(part)
		if !ok {
			continue
		}
		specs = append(specs, spec{
			typ:    typ,
			subtyp: subtyp,
			params: params,
			q:      q,
			order:  i,
			value:  typ + "/" + subtyp,
		})
	}
	return specs
}

func parseMediaRange(s string) (typ, subtyp string, params map[string]string, q float64, ok bool) {
	segs := strings.Split(s, ";")
	full := strings.TrimSpace(segs[0])
	slash := strings.IndexByte(full, '/')
	if slash < 0 {
		return "", "", nil, 0, false
	}
	typ = strings.TrimSpace(full[:slash])
	subtyp = strings.TrimSpace(full[slash+1:])
	if typ == "" || subtyp == "" {
		return "", "", nil, 0, false
	}
	q = 1.0
	params = map[string]string{}
	for _, p := range segs[1:] {
		p = strings.TrimSpace(p)
		eq := strings.IndexByte(p, '=')
		if eq < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(p[:eq]))
		val := strings.Trim(strings.TrimSpace(p[eq+1:]), "\"")
		if key == "q" {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				q = f
			}
		} else {
			params[key] = val
		}
	}
	return typ, subtyp, params, q, true
}

// bestMediaMatch finds the best-matching accept spec for a concrete media type.
func bestMediaMatch(accepts []spec, available string) (spec, bool) {
	atyp, asubtyp, aparams, _, ok := parseMediaRange(available)
	if !ok {
		return spec{}, false
	}
	var best spec
	found := false
	bestScore := -1
	for _, a := range accepts {
		score := 0
		if a.typ == "*" {
			// matches anything
		} else if strings.EqualFold(a.typ, atyp) {
			score += 4
		} else {
			continue
		}
		if a.subtyp == "*" {
			// matches
		} else if strings.EqualFold(a.subtyp, asubtyp) {
			score += 2
		} else {
			continue
		}
		// params must all match
		paramsOK := true
		for k, v := range a.params {
			if av, present := aparams[k]; !present || !strings.EqualFold(av, v) {
				paramsOK = false
				break
			}
		}
		if !paramsOK {
			continue
		}
		score += len(a.params)
		if score > bestScore || (score == bestScore && a.q > best.q) {
			bestScore = score
			best = a
			found = true
		}
	}
	return best, found
}

// ---------------------------------------------------------------------------
// Languages
// ---------------------------------------------------------------------------

// Languages returns the acceptable languages given the request's
// Accept-Language header, restricted to the provided available languages. With
// no arguments it returns all acceptable languages sorted by quality.
func (n *Negotiator) Languages(available ...string) []string {
	accepts := parseSimple(n.header.Get("Accept-Language"), "*")
	if len(available) == 0 {
		return simpleAll(accepts)
	}
	var matched []spec
	for i, av := range available {
		if best, ok := bestLanguageMatch(accepts, av); ok && best.q > 0 {
			best.index = i
			best.value = av
			matched = append(matched, best)
		}
	}
	sortSpecsByMatch(matched)
	return specValues(matched)
}

// Language returns the single best acceptable language, or "".
func (n *Negotiator) Language(available ...string) string {
	res := n.Languages(available...)
	if len(res) == 0 {
		return ""
	}
	return res[0]
}

func bestLanguageMatch(accepts []spec, available string) (spec, bool) {
	var best spec
	found := false
	bestScore := -1
	availLower := strings.ToLower(available)
	availPrefix := availLower
	if dash := strings.IndexByte(availPrefix, '-'); dash >= 0 {
		availPrefix = availPrefix[:dash]
	}
	for _, a := range accepts {
		v := strings.ToLower(a.value)
		score := 0
		if v == "*" {
			score = 0
		} else if v == availLower {
			score = 2
		} else {
			// prefix matching: "en" matches "en-US"
			vPrefix := v
			if dash := strings.IndexByte(vPrefix, '-'); dash >= 0 {
				vPrefix = vPrefix[:dash]
			}
			if vPrefix == availPrefix && (v == vPrefix || availLower == vPrefix) {
				score = 1
			} else {
				continue
			}
		}
		if score > bestScore || (score == bestScore && a.q > best.q) {
			bestScore = score
			best = a
			found = true
		}
	}
	return best, found
}

// ---------------------------------------------------------------------------
// Charsets
// ---------------------------------------------------------------------------

// Charsets returns the acceptable charsets given the request's Accept-Charset
// header, restricted to the provided available charsets. With no arguments it
// returns all acceptable charsets sorted by quality.
func (n *Negotiator) Charsets(available ...string) []string {
	accepts := parseSimple(n.header.Get("Accept-Charset"), "*")
	if len(available) == 0 {
		return simpleAll(accepts)
	}
	var matched []spec
	for i, av := range available {
		if best, ok := bestSimpleMatch(accepts, av); ok && best.q > 0 {
			best.index = i
			best.value = av
			matched = append(matched, best)
		}
	}
	sortSpecsByMatch(matched)
	return specValues(matched)
}

// Charset returns the single best acceptable charset, or "".
func (n *Negotiator) Charset(available ...string) string {
	res := n.Charsets(available...)
	if len(res) == 0 {
		return ""
	}
	return res[0]
}

// ---------------------------------------------------------------------------
// Encodings
// ---------------------------------------------------------------------------

// Encodings returns the acceptable encodings given the request's
// Accept-Encoding header, restricted to the provided available encodings. The
// "identity" encoding is always acceptable unless explicitly disabled. With no
// arguments it returns all acceptable encodings sorted by quality.
func (n *Negotiator) Encodings(available ...string) []string {
	header := n.header.Get("Accept-Encoding")
	accepts := parseSimple(header, "identity")
	// Ensure identity is acceptable unless explicitly excluded.
	if identityQ(accepts) < 0 {
		accepts = append(accepts, spec{value: "identity", q: 0.0001, order: len(accepts)})
	}
	if len(available) == 0 {
		return simpleAll(accepts)
	}
	var matched []spec
	for i, av := range available {
		if best, ok := bestSimpleMatch(accepts, av); ok && best.q > 0 {
			best.index = i
			best.value = av
			matched = append(matched, best)
		}
	}
	sortSpecsByMatch(matched)
	return specValues(matched)
}

// Encoding returns the single best acceptable encoding, or "".
func (n *Negotiator) Encoding(available ...string) string {
	res := n.Encodings(available...)
	if len(res) == 0 {
		return ""
	}
	return res[0]
}

// identityQ returns the quality of an explicit "identity" or "*" entry, or -1.
func identityQ(accepts []spec) float64 {
	for _, a := range accepts {
		if strings.EqualFold(a.value, "identity") || a.value == "*" {
			return a.q
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Shared helpers for simple (single-token) accept headers.
// ---------------------------------------------------------------------------

// parseSimple parses a header consisting of comma-separated tokens with
// optional q-values. def is the default value assumed when the header is empty.
func parseSimple(header, def string) []spec {
	var specs []spec
	if strings.TrimSpace(header) == "" {
		return []spec{{value: def, q: 1.0, order: 0}}
	}
	for i, part := range splitComma(header) {
		segs := strings.Split(part, ";")
		value := strings.TrimSpace(segs[0])
		if value == "" {
			continue
		}
		q := 1.0
		for _, p := range segs[1:] {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(strings.ToLower(p), "q=") {
				if f, err := strconv.ParseFloat(strings.TrimSpace(p[2:]), 64); err == nil {
					q = f
				}
			}
		}
		specs = append(specs, spec{value: value, q: q, order: i})
	}
	return specs
}

func bestSimpleMatch(accepts []spec, available string) (spec, bool) {
	var best spec
	found := false
	bestScore := -1
	for _, a := range accepts {
		score := 0
		if a.value == "*" {
			score = 0
		} else if strings.EqualFold(a.value, available) {
			score = 1
		} else {
			continue
		}
		if score > bestScore || (score == bestScore && a.q > best.q) {
			bestScore = score
			best = a
			found = true
		}
	}
	return best, found
}

func simpleAll(accepts []spec) []string {
	var full []spec
	for _, a := range accepts {
		if a.q > 0 && a.value != "*" {
			full = append(full, a)
		}
	}
	sortSpecs(full)
	return specValues(full)
}

func specValues(specs []spec) []string {
	out := make([]string, 0, len(specs))
	for _, s := range specs {
		out = append(out, s.value)
	}
	return out
}

// sortSpecs sorts by descending q, stable by original order.
func sortSpecs(specs []spec) {
	sort.SliceStable(specs, func(i, j int) bool {
		if specs[i].q != specs[j].q {
			return specs[i].q > specs[j].q
		}
		return specs[i].order < specs[j].order
	})
}

// sortSpecsByMatch sorts matched specs by descending q, stable by the index of
// the available value (preserving the caller's ordering on ties).
func sortSpecsByMatch(specs []spec) {
	sort.SliceStable(specs, func(i, j int) bool {
		if specs[i].q != specs[j].q {
			return specs[i].q > specs[j].q
		}
		return specs[i].index < specs[j].index
	})
}

func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
