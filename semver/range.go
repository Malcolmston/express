package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// semOpSpaceRe matches whitespace directly following a range operator or a
// caret/tilde, so it can be removed (node-semver's comparatorTrimReplace). This
// lets "> 1.0.0", "^ 1", "~ 1.0" tokenize as single comparators.
var semOpSpaceRe = regexp.MustCompile(`(>=|<=|\^|~|>|<|=)\s+`)

// semNormalizeRange collapses "~>" to "~" and strips whitespace after operators.
func semNormalizeRange(s string) string {
	s = strings.ReplaceAll(s, "~>", "~")
	return semOpSpaceRe.ReplaceAllString(s, "$1")
}

// operator is a single version comparator.
type semOp int

const (
	semGT semOp = iota
	semGTE
	semLT
	semLTE
	semEQ
)

type semComparator struct {
	op semOp
	v  *Version
}

func (c semComparator) test(v *Version) bool {
	cmp := v.Compare(c.v)
	switch c.op {
	case semGT:
		return cmp > 0
	case semGTE:
		return cmp >= 0
	case semLT:
		return cmp < 0
	case semLTE:
		return cmp <= 0
	default: // semEQ
		return cmp == 0
	}
}

// Range is a compiled semantic-version range: a disjunction (OR) of tuples,
// where each tuple is a conjunction (AND) of comparators. It matches the common
// npm range grammar (see the package documentation). The zero Range matches
// nothing; construct one with ParseRange.
type Range struct {
	// each element is an AND set of comparators; the Range is satisfied when
	// any one set is fully satisfied.
	sets [][]semComparator
	raw  string
}

// ParseRange compiles a range expression (for example "^1.2.0", ">=1.0.0 <2.0.0",
// "1.x || 2.x", "1.2.3 - 2.3.4"). It returns an error when the expression cannot
// be parsed.
func ParseRange(expr string) (*Range, error) {
	r := &Range{raw: expr}
	for _, orPart := range strings.Split(expr, "||") {
		set, err := semParseComparatorSet(strings.TrimSpace(orPart))
		if err != nil {
			return nil, err
		}
		r.sets = append(r.sets, set)
	}
	if len(r.sets) == 0 {
		return nil, fmt.Errorf("%w: empty range", ErrInvalidVersion)
	}
	return r, nil
}

// String returns the original range expression the Range was parsed from.
func (r *Range) String() string { return r.raw }

// Test reports whether the version satisfies the range. A prerelease version
// only matches when a comparator naming the same Major.Minor.Patch also carries
// a prerelease, mirroring npm's default behaviour.
func (r *Range) Test(v *Version) bool {
	for _, set := range r.sets {
		if semTestSet(set, v) {
			return true
		}
	}
	return false
}

func semTestSet(set []semComparator, v *Version) bool {
	for _, c := range set {
		if !c.test(v) {
			return false
		}
	}
	if len(v.Prerelease) == 0 {
		return true
	}
	// For a prerelease version, require that some comparator names the same
	// core version and itself carries a prerelease.
	for _, c := range set {
		if len(c.v.Prerelease) > 0 &&
			c.v.Major == v.Major && c.v.Minor == v.Minor && c.v.Patch == v.Patch {
			return true
		}
	}
	return false
}

// Satisfies reports whether the version string satisfies the range string. It
// returns false if either input is invalid.
func Satisfies(version, constraint string) bool {
	v, err := Parse(version)
	if err != nil {
		return false
	}
	r, err := ParseRange(constraint)
	if err != nil {
		return false
	}
	return r.Test(v)
}

// MaxSatisfying returns the highest version in versions that satisfies the
// range, or "" (and ok=false) if none do. Invalid version strings are skipped,
// and an unparseable range satisfies nothing. The returned string is the
// matching element exactly as it was given, as npm semver's maxSatisfying
// returns it, so a "v"-prefixed input comes back "v"-prefixed.
func MaxSatisfying(versions []string, constraint string) (string, bool) {
	return semSatisfyingPick(versions, constraint, true)
}

// MinSatisfying returns the lowest version in versions that satisfies the
// range, or "" (and ok=false) if none do. Invalid version strings are skipped,
// and an unparseable range satisfies nothing. Like MaxSatisfying it returns the
// matching element unchanged.
func MinSatisfying(versions []string, constraint string) (string, bool) {
	return semSatisfyingPick(versions, constraint, false)
}

// semSatisfyingPick backs MaxSatisfying and MinSatisfying: the first satisfying
// element wins a tie, so the result is stable in the input order.
func semSatisfyingPick(versions []string, constraint string, max bool) (string, bool) {
	r, err := ParseRange(constraint)
	if err != nil {
		return "", false
	}
	var best *Version
	bestRaw := ""
	for _, s := range versions {
		v, err := Parse(s)
		if err != nil || !r.Test(v) {
			continue
		}
		c := 0
		if best != nil {
			c = v.Compare(best)
		}
		if best == nil || (max && c > 0) || (!max && c < 0) {
			best, bestRaw = v, s
		}
	}
	if best == nil {
		return "", false
	}
	return bestRaw, true
}

// --- range parsing ----------------------------------------------------------

func semParseComparatorSet(s string) ([]semComparator, error) {
	s = semNormalizeRange(strings.TrimSpace(s))
	if s == "" {
		// empty string means "any version"
		return []semComparator{{op: semGTE, v: &Version{}}}, nil
	}
	// Hyphen range: "1.2.3 - 2.3.4"
	if i := strings.Index(s, " - "); i >= 0 {
		return semParseHyphen(strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+3:]))
	}
	fields := strings.Fields(s)
	var out []semComparator
	for _, f := range fields {
		cs, err := semParseSingle(f)
		if err != nil {
			return nil, err
		}
		out = append(out, cs...)
	}
	if len(out) == 0 {
		return []semComparator{{op: semGTE, v: &Version{}}}, nil
	}
	return out, nil
}

func semParseHyphen(lo, hi string) ([]semComparator, error) {
	lp, err := semParsePartial(lo)
	if err != nil {
		return nil, err
	}
	hp, err := semParsePartial(hi)
	if err != nil {
		return nil, err
	}
	// Lower bound: a partial version floors to .0.0, which the cascade in
	// semParsePartial has already done.
	set := []semComparator{{op: semGTE, v: &Version{Major: lp.major, Minor: lp.minor, Patch: lp.patch, Prerelease: lp.pre}}}
	// Upper bound.
	if hp.majorX {
		// hyphen upper bound of "x" imposes no upper limit.
	} else if hp.minorX {
		set = append(set, semComparator{op: semLT, v: &Version{Major: hp.major + 1}})
	} else if hp.patchX {
		set = append(set, semComparator{op: semLT, v: &Version{Major: hp.major, Minor: hp.minor + 1}})
	} else {
		set = append(set, semComparator{op: semLTE, v: &Version{Major: hp.major, Minor: hp.minor, Patch: hp.patch, Prerelease: hp.pre}})
	}
	return set, nil
}

type semPartial struct {
	major, minor, patch    uint64
	majorX, minorX, patchX bool
	pre                    []string

	// The three fields exactly as written, before the cascade above collapses
	// "anything after a wildcard" into a wildcard. minorSet/patchSet say whether
	// the field was present at all, rawMinorX/rawPatchX whether that present
	// field was itself a wildcard. Only invalidXRangeOrder needs this detail.
	minorSet, patchSet   bool
	rawMinorX, rawPatchX bool
}

// invalidXRangeOrder reports whether a numeric field follows a wildcard one,
// as in "1.x.3" or "x.2.3". node-semver's replaceXRange refuses to expand such
// a token, which then fails as an invalid comparator, so the whole range is
// invalid. Caret, tilde and hyphen ranges do not apply this rule: they ignore
// the fields after the first wildcard instead.
func (p semPartial) invalidXRangeOrder() bool {
	if p.majorX && p.minorSet && !p.rawMinorX {
		return true
	}
	minorIsX := p.majorX || !p.minorSet || p.rawMinorX
	return minorIsX && p.patchSet && !p.rawPatchX
}

// semParsePartial parses one range token's version part — node-semver's
// XRANGEPLAIN: an optional run of "v", "=" and whitespace, then one to three
// dot-separated fields each of which is a numeric field without a leading zero
// or a wildcard ("x", "X", "*"), then, only when all three fields are present,
// an optional prerelease and build metadata. Anything else is an error, which is
// how a bogus range ("garbage", ">=1.0.0, <2.0.0", ">=01.2.3", "^latest") comes
// to be rejected rather than silently treated as a wildcard.
func semParsePartial(s string) (semPartial, error) {
	p := semPartial{}
	bad := func(why string) (semPartial, error) {
		return semPartial{}, fmt.Errorf("%w: range %q: %s", ErrInvalidVersion, s, why)
	}
	// [v=\s]* — the same prefix run node-semver's XRANGEPLAIN allows.
	rest := strings.TrimLeft(s, "v= \t")
	if rest == "" {
		return bad("empty comparator")
	}

	// field reads one wildcard-or-numeric field and reports whether it was a
	// wildcard.
	field := func(tok string) (uint64, bool, error) {
		switch tok {
		case "x", "X", "*":
			return 0, true, nil
		}
		if !semIsNumericID(tok) {
			return 0, false, fmt.Errorf("%w: range %q: %q is not a version field", ErrInvalidVersion, s, tok)
		}
		if len(tok) > 1 && tok[0] == '0' {
			return 0, false, fmt.Errorf("%w: range %q: field %q has a leading zero", ErrInvalidVersion, s, tok)
		}
		n, err := strconv.ParseUint(tok, 10, 64)
		if err != nil || n > semMaxSafeInteger {
			return 0, false, fmt.Errorf("%w: range %q: field %q is too big", ErrInvalidVersion, s, tok)
		}
		return n, false, nil
	}

	// Split off the prerelease and build metadata, which only the three-field
	// form may carry; the check that all three fields are present happens below.
	core := rest
	var pre, build []string
	if i := strings.IndexByte(core, '+'); i >= 0 {
		b := core[i+1:]
		core = core[:i]
		if b == "" {
			return bad("empty build metadata")
		}
		build = strings.Split(b, ".")
		for _, id := range build {
			if !semIsBuildID(id) {
				return bad("invalid build identifier " + strconv.Quote(id))
			}
		}
	}
	if i := strings.IndexByte(core, '-'); i >= 0 {
		pr := core[i+1:]
		core = core[:i]
		if pr == "" {
			return bad("empty prerelease")
		}
		pre = strings.Split(pr, ".")
		for _, id := range pre {
			if !semIsPrereleaseID(id) {
				return bad("invalid prerelease identifier " + strconv.Quote(id))
			}
		}
	}

	parts := strings.Split(core, ".")
	if len(parts) > 3 {
		return bad("more than three version fields")
	}
	if (len(pre) > 0 || len(build) > 0) && len(parts) != 3 {
		return bad("a partial version may not carry a prerelease or build metadata")
	}
	var err error
	if p.major, p.majorX, err = field(parts[0]); err != nil {
		return semPartial{}, err
	}
	if len(parts) > 1 {
		p.minorSet = true
		if p.minor, p.rawMinorX, err = field(parts[1]); err != nil {
			return semPartial{}, err
		}
	}
	if len(parts) > 2 {
		p.patchSet = true
		if p.patch, p.rawPatchX, err = field(parts[2]); err != nil {
			return semPartial{}, err
		}
	}
	p.pre = pre

	// Cascade: once a field is a wildcard (or absent) every lower field is one
	// too, and its numeric value is meaningless.
	p.minorX = p.majorX || !p.minorSet || p.rawMinorX
	p.patchX = p.minorX || !p.patchSet || p.rawPatchX
	if p.majorX {
		p.major, p.minor, p.patch = 0, 0, 0
	} else if p.minorX {
		p.minor, p.patch = 0, 0
	} else if p.patchX {
		p.patch = 0
	}
	return p, nil
}

func semParseSingle(f string) ([]semComparator, error) {
	// caret and tilde
	if strings.HasPrefix(f, "^") {
		return semParseCaret(f[1:])
	}
	if strings.HasPrefix(f, "~") {
		return semParseTilde(f[1:])
	}
	// operators
	for _, opStr := range []string{">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(f, opStr) {
			rest := strings.TrimSpace(f[len(opStr):])
			return semParseOpVersion(opStr, rest)
		}
	}
	// bare version or x-range
	return semParseXRange(f)
}

func semParseOpVersion(opStr, rest string) ([]semComparator, error) {
	op := map[string]semOp{">=": semGTE, "<=": semLTE, ">": semGT, "<": semLT, "=": semEQ}[opStr]
	p, err := semParsePartial(rest)
	if err != nil {
		return nil, err
	}
	if p.invalidXRangeOrder() {
		return nil, fmt.Errorf("%w: range %q%q: a numeric field follows a wildcard", ErrInvalidVersion, opStr, rest)
	}
	anyX := p.majorX || p.minorX || p.patchX

	// An operator applied to a partial ("x-range") version expands per
	// node-semver's replaceXRange, e.g. ">1" -> ">=2.0.0", "<=0.7.x" -> "<0.8.0",
	// "=0.7.x" -> ">=0.7.0 <0.8.0". A fully specified version keeps the operator.
	if !anyX {
		v := &Version{Major: p.major, Minor: p.minor, Patch: p.patch, Prerelease: p.pre}
		return []semComparator{{op: op, v: v}}, nil
	}

	// "=" with any wildcard degrades to a plain x-range.
	if op == semEQ {
		return semXRangeSet(p), nil
	}

	if p.majorX {
		// ">X"/"<X" match nothing; ">=X"/"<=X" match anything.
		if op == semGT || op == semLT {
			return []semComparator{{op: semLT, v: &Version{}}}, nil
		}
		return []semComparator{{op: semGTE, v: &Version{}}}, nil
	}

	// minor or patch is a wildcard.
	switch op {
	case semGT:
		// >1 => >=2.0.0 ; >1.2 => >=1.3.0
		if p.minorX {
			return []semComparator{{op: semGTE, v: &Version{Major: p.major + 1}}}, nil
		}
		return []semComparator{{op: semGTE, v: &Version{Major: p.major, Minor: p.minor + 1}}}, nil
	case semLTE:
		// <=1 => <2.0.0 ; <=1.2 => <1.3.0
		if p.minorX {
			return []semComparator{{op: semLT, v: &Version{Major: p.major + 1}}}, nil
		}
		return []semComparator{{op: semLT, v: &Version{Major: p.major, Minor: p.minor + 1}}}, nil
	case semLT:
		// <1 => <1.0.0 ; <1.2 => <1.2.0
		return []semComparator{{op: semLT, v: &Version{Major: p.major, Minor: p.minor}}}, nil
	default: // semGTE
		// >=1 => >=1.0.0 ; >=1.2 => >=1.2.0
		return []semComparator{{op: semGTE, v: &Version{Major: p.major, Minor: p.minor}}}, nil
	}
}

// semXRangeSet expands a partial ("x-range") version into comparators.
func semXRangeSet(p semPartial) []semComparator {
	if p.majorX {
		return []semComparator{{op: semGTE, v: &Version{}}}
	}
	if p.minorX {
		// e.g. "1" or "1.x" -> >=1.0.0 <2.0.0
		return []semComparator{
			{op: semGTE, v: &Version{Major: p.major}},
			{op: semLT, v: &Version{Major: p.major + 1}},
		}
	}
	if p.patchX {
		// e.g. "1.2" or "1.2.x" -> >=1.2.0 <1.3.0
		return []semComparator{
			{op: semGTE, v: &Version{Major: p.major, Minor: p.minor}},
			{op: semLT, v: &Version{Major: p.major, Minor: p.minor + 1}},
		}
	}
	// fully specified: exact match
	return []semComparator{{op: semEQ, v: &Version{Major: p.major, Minor: p.minor, Patch: p.patch, Prerelease: p.pre}}}
}

func semParseXRange(f string) ([]semComparator, error) {
	if f == "" || f == "*" || f == "x" || f == "X" {
		return []semComparator{{op: semGTE, v: &Version{}}}, nil
	}
	p, err := semParsePartial(f)
	if err != nil {
		return nil, err
	}
	if p.invalidXRangeOrder() {
		return nil, fmt.Errorf("%w: range %q: a numeric field follows a wildcard", ErrInvalidVersion, f)
	}
	return semXRangeSet(p), nil
}

func semParseCaret(rest string) ([]semComparator, error) {
	// Caret, tilde and hyphen ranges do not reject "1.x.3": node-semver expands
	// them from the first wildcard on and ignores the rest.
	p, err := semParsePartial(rest)
	if err != nil {
		return nil, err
	}
	if p.majorX {
		// "^x"/"^*" impose no constraint.
		return []semComparator{{op: semGTE, v: &Version{}}}, nil
	}
	lo := &Version{Major: p.major, Minor: p.minor, Patch: p.patch, Prerelease: p.pre}
	var hi *Version
	switch {
	case p.major > 0 || p.minorX:
		hi = &Version{Major: p.major + 1}
	case p.minor > 0 || p.patchX:
		hi = &Version{Major: 0, Minor: p.minor + 1}
	default:
		hi = &Version{Major: 0, Minor: 0, Patch: p.patch + 1}
	}
	return []semComparator{{op: semGTE, v: lo}, {op: semLT, v: hi}}, nil
}

func semParseTilde(rest string) ([]semComparator, error) {
	p, err := semParsePartial(rest)
	if err != nil {
		return nil, err
	}
	if p.majorX {
		// "~x"/"~*" impose no constraint.
		return []semComparator{{op: semGTE, v: &Version{}}}, nil
	}
	lo := &Version{Major: p.major, Minor: p.minor, Patch: p.patch, Prerelease: p.pre}
	var hi *Version
	if p.minorX {
		// ~1 -> >=1.0.0 <2.0.0
		hi = &Version{Major: p.major + 1}
	} else {
		// ~1.2 or ~1.2.3 -> >=... <1.3.0
		hi = &Version{Major: p.major, Minor: p.minor + 1}
	}
	return []semComparator{{op: semGTE, v: lo}, {op: semLT, v: hi}}, nil
}
