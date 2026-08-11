// Package negotiator is an HTTP content negotiation helper, a stdlib-only Go
// port of the npm "negotiator" package (https://www.npmjs.com/package/negotiator).
// It inspects request headers such as Accept, Accept-Language, Accept-Charset
// and Accept-Encoding and reports which of a set of server-side representations
// best match the client's stated preferences. Like the Node original, it is a
// low-level building block: it makes no assumptions about your routes and only
// answers the question "given these headers and these available options, which
// options are acceptable and in what order?".
//
// Content negotiation matters whenever a single URL can be served in more than
// one form. A client advertises what it can handle by sending weighted lists in
// its request headers, for example "Accept: application/json;q=0.9, text/html"
// or "Accept-Encoding: gzip, br;q=0.8". The server holds a fixed menu of things
// it is able to produce (JSON or HTML, gzip or brotli, English or French) and
// must pick the best mutually agreeable choice. This package performs that pick,
// mirroring the behavior that Express, the accepts package, and many other Node
// middlewares rely on internally.
//
// A Negotiator is constructed from an http.Header via New and exposes four
// families of methods: MediaTypes/MediaType for the Accept header,
// Languages/Language for Accept-Language, Charsets/Charset for Accept-Charset,
// and Encodings/Encoding for Accept-Encoding. The plural methods return every
// acceptable option in descending preference order; the singular methods return
// just the best option, or the empty string when nothing is acceptable. Each
// method optionally takes the list of available options; called with no
// arguments the plural forms instead report the client's own preferences parsed
// straight from the header (with q=0 entries omitted).
//
// Preference is driven by the quality value ("q") attached to each header
// entry, a float in the range 0 through 1 that defaults to 1 when omitted. A
// higher q means a stronger preference, and a q of 0 explicitly rejects an
// option: such options are never returned even if they would otherwise match a
// wildcard. A q that is not a number at all (for example "q=high") makes the
// whole entry unacceptable, exactly as the Node original's parseFloat produces
// NaN and NaN > 0 is false. Matching also accounts for specificity — for media
// types a concrete "text/html" outranks "text/*" which outranks "*/*", and
// parameters on an accept range (such as level or profile) must all be present
// on the candidate for the range to apply. Specificity is compared before
// quality when choosing which header entry applies to a candidate, and quality
// is compared before specificity when ordering the results, both as upstream
// does. When candidates tie completely the original order is preserved, giving
// the server the final say among equally-weighted choices.
//
// A few edge cases match the Node semantics deliberately. An *absent* Accept
// header is treated as "*/*", so all offered media types are acceptable, while a
// header that is present but empty accepts nothing at all — the same split
// applies to Accept-Language ("*") and Accept-Charset ("*"). Language matching
// is prefix-aware, so a header of "en" matches an available "en-US" (and vice
// versa) at a lower specificity than an exact tag match. Encodings always treat
// "identity" (no transformation) as acceptable unless it is explicitly disabled
// with identity;q=0 or a "*;q=0" catch-all, which is why a request advertising
// only "gzip" will still permit an uncompressed response; the implicit identity
// is injected at the lowest quality seen in the header. This port relies solely
// on the Go standard library and returns plain string slices, so it composes
// naturally with net/http handlers.
package negotiator

import (
	"math"
	"net/http"
	"regexp"
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

// field returns the raw header value and whether the field was sent at all.
// Upstream reads request.headers[name], so an absent field is undefined (and
// falls back to a wildcard) while a field present with an empty value is the
// empty string (and accepts nothing). Keeping the two apart is the only way to
// reproduce that.
func (n *Negotiator) field(name string) (value string, present bool) {
	vs := n.header.Values(name)
	if len(vs) == 0 {
		return "", false
	}
	return n.header.Get(name), true
}

// ---------------------------------------------------------------------------
// JavaScript arithmetic emulation
//
// Upstream's comparators chain numeric subtractions with `||`, which treats
// both 0 and NaN as falsy, and its q-values come from parseFloat, which yields
// NaN for anything unparseable. Reproducing the ordering therefore means
// reproducing those two behaviours exactly.
// ---------------------------------------------------------------------------

// falsy reports whether a number is falsy in JavaScript's sense: zero (either
// sign) or NaN.
func falsy(f float64) bool { return f == 0 || math.IsNaN(f) }

var jsNumberRE = regexp.MustCompile(`^[+-]?(?:Infinity|(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?)`)

// parseFloatJS mimics JavaScript's global parseFloat: leading whitespace is
// skipped, the longest numeric prefix is used, and anything else (including the
// empty string and a missing value) is NaN.
func parseFloatJS(s string, present bool) float64 {
	if !present {
		return math.NaN()
	}
	s = strings.TrimLeft(s, " \t\n\r\f\v ")
	m := jsNumberRE.FindString(s)
	if m == "" {
		return math.NaN()
	}
	if strings.HasSuffix(m, "Infinity") {
		if strings.HasPrefix(m, "-") {
			return math.Inf(-1)
		}
		return math.Inf(1)
	}
	f, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return math.NaN()
	}
	return f
}

// priority is the {i, o, q, s} record upstream builds for each candidate: i is
// the candidate's index in the available list, o the index of the header entry
// that matched it, q the quality of that entry and s the specificity of the
// match.
type priority struct {
	i, o int
	q    float64
	s    int
	// value is the string reported for this priority.
	value string
}

// supersedes reports whether cand should replace cur while scanning header
// entries, reproducing upstream's
// `(priority.s - spec.s || priority.q - spec.q || priority.o - spec.o) < 0`.
func supersedes(cur, cand priority) bool {
	d := float64(cur.s - cand.s)
	if falsy(d) {
		d = cur.q - cand.q
	}
	if falsy(d) {
		d = float64(cur.o - cand.o)
	}
	// NaN < 0 is false in both languages, so an unparseable quality simply
	// fails to win the comparison at that step.
	return d < 0
}

// sortPriorities applies upstream's compareSpecs:
// `(b.q - a.q) || (b.s - a.s) || (a.o - b.o) || (a.i - b.i) || 0`.
func sortPriorities(ps []priority) {
	sort.SliceStable(ps, func(x, y int) bool {
		a, b := ps[x], ps[y]
		d := b.q - a.q
		if falsy(d) {
			d = float64(b.s - a.s)
		}
		if falsy(d) {
			d = float64(a.o - b.o)
		}
		if falsy(d) {
			d = float64(a.i - b.i)
		}
		return d < 0
	})
}

// sortParsed orders the header's own entries for the no-arguments form. The
// parse results carry no `s` or `o` field, so upstream's compareSpecs falls
// through NaN comparisons and effectively sorts by quality then header order.
func sortParsed(ps []priority) {
	sort.SliceStable(ps, func(x, y int) bool {
		a, b := ps[x], ps[y]
		d := b.q - a.q
		if falsy(d) {
			d = float64(a.i - b.i)
		}
		return d < 0
	})
}

// acceptable filters out entries whose quality is not strictly positive, which
// also removes NaN qualities.
func acceptable(ps []priority) []priority {
	out := ps[:0:0]
	for _, p := range ps {
		if p.q > 0 {
			out = append(out, p)
		}
	}
	return out
}

func values(ps []priority) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.value)
	}
	return out
}

// quoteCount counts double quotes, used to keep a comma or semicolon inside a
// quoted parameter value from splitting the list.
func quoteCount(s string) int { return strings.Count(s, `"`) }

// splitKeepingQuotes splits on sep but rejoins a piece onto the previous one
// while that previous piece has an odd number of quotes.
func splitKeepingQuotes(s, sep string) []string {
	raw := strings.Split(s, sep)
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		if n := len(out); n > 0 && quoteCount(out[n-1])%2 == 1 {
			out[n-1] += sep + part
		} else {
			out = append(out, part)
		}
	}
	return out
}

// splitKeyValuePair splits at the first '=' and reports whether a value was
// present at all (upstream leaves it undefined when there is no '=').
func splitKeyValuePair(s string) (key, val string, hasVal bool) {
	if i := strings.IndexByte(s, '='); i >= 0 {
		return s[:i], s[i+1:], true
	}
	return s, "", false
}

// ---------------------------------------------------------------------------
// Media types (Accept)
// ---------------------------------------------------------------------------

var simpleMediaTypeRE = regexp.MustCompile(`^[ \t\n\r\f\v]*([^ \t\n\r\f\v/;]+)/([^;\s]+)[ \t\n\r\f\v]*(?:;(.*))?$`)

type mediaSpec struct {
	typ, subtype string
	params       map[string]string
	q            float64
	i            int
}

func parseMediaType(str string, i int) (mediaSpec, bool) {
	m := simpleMediaTypeRE.FindStringSubmatch(str)
	if m == nil {
		return mediaSpec{}, false
	}
	spec := mediaSpec{typ: m[1], subtype: m[2], params: map[string]string{}, q: 1, i: i}
	if m[3] != "" {
		for _, kvp := range splitKeepingQuotes(m[3], ";") {
			key, val, hasVal := splitKeyValuePair(strings.TrimSpace(kvp))
			// A quoted value is unwrapped; a missing value stays empty, which
			// is how `undefined || ''` behaves upstream.
			if len(val) > 1 && val[0] == '"' && val[len(val)-1] == '"' {
				val = val[1 : len(val)-1]
			}
			if strings.ToLower(key) == "q" {
				spec.q = parseFloatJS(val, hasVal)
				// Everything after q is an accept-extension, not part of the
				// media type for matching purposes.
				break
			}
			spec.params[strings.ToLower(key)] = val
		}
	}
	return spec, true
}

func parseAccept(accept string) []mediaSpec {
	var out []mediaSpec
	for i, part := range splitKeepingQuotes(accept, ",") {
		if spec, ok := parseMediaType(strings.TrimSpace(part), i); ok {
			out = append(out, spec)
		}
	}
	return out
}

// specifyMedia is upstream's specify(): it reports the specificity with which
// an accept entry covers a concrete media type, or false when it does not.
func specifyMedia(typ string, spec mediaSpec, index int) (priority, bool) {
	p, ok := parseMediaType(typ, 0)
	if !ok {
		return priority{}, false
	}
	s := 0
	switch {
	case strings.EqualFold(spec.typ, p.typ):
		s |= 4
	case spec.typ != "*":
		return priority{}, false
	}
	switch {
	case strings.EqualFold(spec.subtype, p.subtype):
		s |= 2
	case spec.subtype != "*":
		return priority{}, false
	}
	if len(spec.params) > 0 {
		for k, v := range spec.params {
			if v == "*" || strings.EqualFold(v, p.params[k]) {
				continue
			}
			return priority{}, false
		}
		s |= 1
	}
	return priority{i: index, o: spec.i, q: spec.q, s: s}, true
}

func mediaTypePriority(typ string, accepted []mediaSpec, index int) priority {
	best := priority{o: -1, q: 0, s: 0, i: index}
	for _, spec := range accepted {
		cand, ok := specifyMedia(typ, spec, index)
		if ok && supersedes(best, cand) {
			best = cand
		}
	}
	best.i = index
	best.value = typ
	return best
}

// MediaTypes returns the media types that are acceptable given the request's
// Accept header, restricted to the provided available media types. When called
// with no arguments it returns every acceptable media type named in the header
// itself, sorted by descending quality.
func (n *Negotiator) MediaTypes(available ...string) []string {
	raw, present := n.field("Accept")
	if !present {
		raw = "*/*"
	}
	accepts := parseAccept(raw)

	if len(available) == 0 {
		ps := make([]priority, 0, len(accepts))
		for _, spec := range accepts {
			ps = append(ps, priority{q: spec.q, i: spec.i, value: spec.typ + "/" + spec.subtype})
		}
		ps = acceptable(ps)
		sortParsed(ps)
		return values(ps)
	}

	ps := make([]priority, 0, len(available))
	for i, av := range available {
		ps = append(ps, mediaTypePriority(av, accepts, i))
	}
	ps = acceptable(ps)
	sortPriorities(ps)
	return values(ps)
}

// MediaType returns the single best acceptable media type, or "" if none of the
// available media types are acceptable.
func (n *Negotiator) MediaType(available ...string) string {
	return first(n.MediaTypes(available...))
}

// ---------------------------------------------------------------------------
// Languages (Accept-Language)
// ---------------------------------------------------------------------------

var simpleLanguageRE = regexp.MustCompile(`^[ \t\n\r\f\v]*([^ \t\n\r\f\v;-]+)(?:-([^;\s]+))?[ \t\n\r\f\v]*(?:;(.*))?$`)

type langSpec struct {
	prefix, suffix, full string
	q                    float64
	i                    int
}

func parseLanguage(str string, i int) (langSpec, bool) {
	m := simpleLanguageRE.FindStringSubmatch(str)
	if m == nil {
		return langSpec{}, false
	}
	spec := langSpec{prefix: m[1], suffix: m[2], full: m[1], q: 1, i: i}
	if m[2] != "" {
		spec.full += "-" + m[2]
	}
	if m[3] != "" {
		// Upstream does not trim these parameter segments, so only a q that
		// directly follows the ';' is recognised — "en; q=0.5" keeps q=1.
		for _, param := range strings.Split(m[3], ";") {
			key, val, hasVal := splitKeyValuePair(param)
			if key == "q" {
				spec.q = parseFloatJS(val, hasVal)
			}
		}
	}
	return spec, true
}

func parseAcceptLanguage(accept string) []langSpec {
	var out []langSpec
	for i, part := range strings.Split(accept, ",") {
		if spec, ok := parseLanguage(strings.TrimSpace(part), i); ok {
			out = append(out, spec)
		}
	}
	return out
}

func specifyLanguage(language string, spec langSpec, index int) (priority, bool) {
	p, ok := parseLanguage(language, 0)
	if !ok {
		return priority{}, false
	}
	s := 0
	switch {
	case strings.EqualFold(spec.full, p.full):
		s |= 4
	case strings.EqualFold(spec.prefix, p.full):
		s |= 2
	case strings.EqualFold(spec.full, p.prefix):
		s |= 1
	case spec.full != "*":
		return priority{}, false
	}
	return priority{i: index, o: spec.i, q: spec.q, s: s}, true
}

// Languages returns the acceptable languages given the request's
// Accept-Language header, restricted to the provided available languages. With
// no arguments it returns all acceptable languages sorted by quality.
func (n *Negotiator) Languages(available ...string) []string {
	raw, present := n.field("Accept-Language")
	if !present {
		raw = "*"
	}
	accepts := parseAcceptLanguage(raw)

	if len(available) == 0 {
		ps := make([]priority, 0, len(accepts))
		for _, spec := range accepts {
			ps = append(ps, priority{q: spec.q, i: spec.i, value: spec.full})
		}
		ps = acceptable(ps)
		sortParsed(ps)
		return values(ps)
	}

	ps := make([]priority, 0, len(available))
	for i, av := range available {
		best := priority{o: -1, q: 0, s: 0, i: i}
		for _, spec := range accepts {
			if cand, ok := specifyLanguage(av, spec, i); ok && supersedes(best, cand) {
				best = cand
			}
		}
		best.i = i
		best.value = av
		ps = append(ps, best)
	}
	ps = acceptable(ps)
	sortPriorities(ps)
	return values(ps)
}

// Language returns the single best acceptable language, or "".
func (n *Negotiator) Language(available ...string) string {
	return first(n.Languages(available...))
}

// ---------------------------------------------------------------------------
// Charsets and encodings (single-token headers)
// ---------------------------------------------------------------------------

var simpleTokenRE = regexp.MustCompile(`^[ \t\n\r\f\v]*([^;\s]+)[ \t\n\r\f\v]*(?:;(.*))?$`)

type tokenSpec struct {
	value string
	q     float64
	i     int
}

func parseToken(str string, i int) (tokenSpec, bool) {
	m := simpleTokenRE.FindStringSubmatch(str)
	if m == nil {
		return tokenSpec{}, false
	}
	spec := tokenSpec{value: m[1], q: 1, i: i}
	if m[2] != "" {
		for _, param := range strings.Split(m[2], ";") {
			key, val, hasVal := splitKeyValuePair(strings.TrimSpace(param))
			if key == "q" {
				spec.q = parseFloatJS(val, hasVal)
				break
			}
		}
	}
	return spec, true
}

func parseTokenList(accept string) []tokenSpec {
	var out []tokenSpec
	for i, part := range strings.Split(accept, ",") {
		if spec, ok := parseToken(strings.TrimSpace(part), i); ok {
			out = append(out, spec)
		}
	}
	return out
}

// specifyToken is the shared specify() for charsets and encodings: an exact
// case-insensitive hit scores 1, the literal "*" scores 0, anything else does
// not match.
func specifyToken(value string, spec tokenSpec, index int) (priority, bool) {
	s := 0
	switch {
	case strings.EqualFold(spec.value, value):
		s |= 1
	case spec.value != "*":
		return priority{}, false
	}
	return priority{i: index, o: spec.i, q: spec.q, s: s}, true
}

func negotiateTokens(accepts []tokenSpec, available []string) []string {
	ps := make([]priority, 0, len(available))
	for i, av := range available {
		best := priority{o: -1, q: 0, s: 0, i: i}
		for _, spec := range accepts {
			if cand, ok := specifyToken(av, spec, i); ok && supersedes(best, cand) {
				best = cand
			}
		}
		best.i = i
		best.value = av
		ps = append(ps, best)
	}
	ps = acceptable(ps)
	sortPriorities(ps)
	return values(ps)
}

func allTokens(accepts []tokenSpec) []string {
	ps := make([]priority, 0, len(accepts))
	for _, spec := range accepts {
		ps = append(ps, priority{q: spec.q, i: spec.i, value: spec.value})
	}
	ps = acceptable(ps)
	sortParsed(ps)
	return values(ps)
}

// Charsets returns the acceptable charsets given the request's Accept-Charset
// header, restricted to the provided available charsets. With no arguments it
// returns all acceptable charsets sorted by quality.
func (n *Negotiator) Charsets(available ...string) []string {
	raw, present := n.field("Accept-Charset")
	if !present {
		raw = "*"
	}
	accepts := parseTokenList(raw)
	if len(available) == 0 {
		return allTokens(accepts)
	}
	return negotiateTokens(accepts, available)
}

// Charset returns the single best acceptable charset, or "".
func (n *Negotiator) Charset(available ...string) string {
	return first(n.Charsets(available...))
}

// parseAcceptEncoding parses the Accept-Encoding header and, when neither
// "identity" nor "*" appears, appends an implicit identity entry at the lowest
// quality seen — the "identity is always acceptable unless refused" rule. Note
// that an absent header is *not* special-cased here: upstream passes
// `accept || ”`, so absent and empty behave identically and only the implicit
// identity survives.
func parseAcceptEncoding(accept string) []tokenSpec {
	parts := strings.Split(accept, ",")
	var out []tokenSpec
	hasIdentity := false
	minQuality := 1.0
	for i, part := range parts {
		spec, ok := parseToken(strings.TrimSpace(part), i)
		if !ok {
			continue
		}
		out = append(out, spec)
		if strings.EqualFold(spec.value, "identity") || spec.value == "*" {
			hasIdentity = true
		}
		// `encoding.q || 1` treats both 0 and NaN as 1, so a refusal does not
		// drag the implicit identity's quality down.
		q := spec.q
		if falsy(q) {
			q = 1
		}
		minQuality = math.Min(minQuality, q)
	}
	if !hasIdentity {
		out = append(out, tokenSpec{value: "identity", q: minQuality, i: len(parts)})
	}
	return out
}

// Encodings returns the acceptable encodings given the request's
// Accept-Encoding header, restricted to the provided available encodings. The
// "identity" encoding is always acceptable unless explicitly disabled. With no
// arguments it returns all acceptable encodings sorted by quality.
func (n *Negotiator) Encodings(available ...string) []string {
	raw, _ := n.field("Accept-Encoding")
	accepts := parseAcceptEncoding(raw)
	if len(available) == 0 {
		return allTokens(accepts)
	}
	return negotiateTokens(accepts, available)
}

// Encoding returns the single best acceptable encoding, or "".
func (n *Negotiator) Encoding(available ...string) string {
	return first(n.Encodings(available...))
}

func first(vs []string) string {
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}
