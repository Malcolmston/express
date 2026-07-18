package semver

import (
	"fmt"
	"strconv"
	"strings"
)

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
// range, or "" (and ok=false) if none do. Invalid version strings are skipped.
func MaxSatisfying(versions []string, constraint string) (string, bool) {
	r, err := ParseRange(constraint)
	if err != nil {
		return "", false
	}
	var best *Version
	for _, s := range versions {
		v, err := Parse(s)
		if err != nil || !r.Test(v) {
			continue
		}
		if best == nil || v.Compare(best) > 0 {
			best = v
		}
	}
	if best == nil {
		return "", false
	}
	return best.String(), true
}

// MinSatisfying returns the lowest version in versions that satisfies the
// range, or "" (and ok=false) if none do. Invalid version strings are skipped.
func MinSatisfying(versions []string, constraint string) (string, bool) {
	r, err := ParseRange(constraint)
	if err != nil {
		return "", false
	}
	var best *Version
	for _, s := range versions {
		v, err := Parse(s)
		if err != nil || !r.Test(v) {
			continue
		}
		if best == nil || v.Compare(best) < 0 {
			best = v
		}
	}
	if best == nil {
		return "", false
	}
	return best.String(), true
}

// --- range parsing ----------------------------------------------------------

func semParseComparatorSet(s string) ([]semComparator, error) {
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
	lv, err := semCoercePartial(lo, false)
	if err != nil {
		return nil, err
	}
	// lower bound: if partial, floor to .0.0 already done by semCoercePartial
	set := []semComparator{{op: semGTE, v: lv}}
	// upper bound
	hp := semSplitPartial(hi)
	if hp.minorX {
		set = append(set, semComparator{op: semLT, v: &Version{Major: hp.major + 1}})
	} else if hp.patchX {
		set = append(set, semComparator{op: semLT, v: &Version{Major: hp.major, Minor: hp.minor + 1}})
	} else {
		set = append(set, semComparator{op: semLTE, v: &Version{Major: hp.major, Minor: hp.minor, Patch: hp.patch, Prerelease: hp.pre}})
	}
	return set, nil
}

type semPartial struct {
	major, minor, patch uint64
	minorX, patchX      bool
	pre                 []string
}

func semSplitPartial(s string) semPartial {
	p := semPartial{}
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	// strip build
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '-'); i >= 0 {
		p.pre = strings.Split(s[i+1:], ".")
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	get := func(idx int) (uint64, bool) {
		if idx >= len(parts) {
			return 0, true // missing -> x
		}
		t := parts[idx]
		if t == "" || t == "x" || t == "X" || t == "*" {
			return 0, true
		}
		n, err := strconv.ParseUint(t, 10, 64)
		if err != nil {
			return 0, true
		}
		return n, false
	}
	var xr bool
	p.major, xr = get(0)
	if xr {
		p.minorX = true
		p.patchX = true
		return p
	}
	p.minor, xr = get(1)
	if xr {
		p.minorX = true
		p.patchX = true
		return p
	}
	p.patch, xr = get(2)
	if xr {
		p.patchX = true
	}
	return p
}

// semCoercePartial turns a partial version into a concrete lower-bound Version.
func semCoercePartial(s string, upper bool) (*Version, error) {
	p := semSplitPartial(s)
	return &Version{Major: p.major, Minor: p.minor, Patch: p.patch, Prerelease: p.pre}, nil
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
	p := semSplitPartial(rest)
	v := &Version{Major: p.major, Minor: p.minor, Patch: p.patch, Prerelease: p.pre}
	// For partial versions with operators, npm expands but we approximate with
	// the floored version, which is correct for the common concrete cases.
	return []semComparator{{op: op, v: v}}, nil
}

func semParseXRange(f string) ([]semComparator, error) {
	if f == "" || f == "*" || f == "x" || f == "X" {
		return []semComparator{{op: semGTE, v: &Version{}}}, nil
	}
	p := semSplitPartial(f)
	if p.minorX {
		// e.g. "1" or "1.x" -> >=1.0.0 <2.0.0 ; but "*" handled above
		return []semComparator{
			{op: semGTE, v: &Version{Major: p.major}},
			{op: semLT, v: &Version{Major: p.major + 1}},
		}, nil
	}
	if p.patchX {
		// e.g. "1.2" or "1.2.x" -> >=1.2.0 <1.3.0
		return []semComparator{
			{op: semGTE, v: &Version{Major: p.major, Minor: p.minor}},
			{op: semLT, v: &Version{Major: p.major, Minor: p.minor + 1}},
		}, nil
	}
	// fully specified: exact match
	return []semComparator{{op: semEQ, v: &Version{Major: p.major, Minor: p.minor, Patch: p.patch, Prerelease: p.pre}}}, nil
}

func semParseCaret(rest string) ([]semComparator, error) {
	p := semSplitPartial(rest)
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
	p := semSplitPartial(rest)
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
