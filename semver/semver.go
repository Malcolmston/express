// Package semver is a standard-library-only Go port of the npm "semver"
// package (https://www.npmjs.com/package/semver), the reference implementation
// of Semantic Versioning 2.0.0 (https://semver.org) used throughout the Node
// and Express ecosystems. It parses, compares, increments, coerces and
// range-matches version strings with the same semantics JavaScript projects
// rely on, so tooling that reasons about dependency versions can be ported to
// Go without pulling in a third-party dependency.
//
// A version is Major.Minor.Patch with optional dot-separated prerelease and
// build-metadata identifiers, for example "1.2.3-alpha.1+build.5". Parse turns
// a string into a *Version, accepting surrounding whitespace and a single
// optional lowercase "v" exactly as node-semver's strict grammar does — an
// uppercase "V" or a leading "=" is loose-mode-only and is rejected. MustParse
// panics instead of returning an error and is convenient for constants and
// tests. Valid reports whether a string is a well-formed version, and Clean is
// the more permissive normalizer: it trims whitespace and strips any run of
// leading "=" and "v" characters before parsing.
//
// Ordering follows the spec precisely: numeric fields compare numerically, a
// version with a prerelease has lower precedence than the same version without
// one, prerelease identifiers compare left to right (numeric identifiers
// numerically, alphanumeric identifiers lexically in ASCII order, and numeric
// identifiers always rank below alphanumeric ones), and build metadata is
// ignored for precedence. Compare returns -1, 0 or +1; the boolean helpers GT,
// GTE, LT, LTE, EQ and NEQ wrap it, and Sort orders a slice ascending.
//
// Inc and the IncMajor/IncMinor/IncPatch methods produce the next version for a
// release level, resetting the lower fields and clearing prerelease and build
// metadata exactly as npm's semver does. Coerce extracts the first
// version-shaped run of digits from arbitrary text ("v2 release" -> "2.0.0").
//
// Range matching is provided by Satisfies and the Range type, which understand
// the common npm range grammar: plain comparators (">=1.2.0", "<2.0.0", "=1.0.0"
// and a bare "1.2.3"), caret ranges ("^1.2.3"), tilde ranges ("~1.2.0"),
// x-ranges ("1.x", "1.2.*", "*"), hyphen ranges ("1.2.3 - 2.3.4"),
// whitespace-separated AND terms within a set (npm has no comma AND), and "||"
// separating alternative sets. Prerelease versions only satisfy a range when a
// comparator in the same tuple names the same Major.Minor.Patch with its own
// prerelease, matching npm's default (includePrerelease=false) behaviour. A
// range expression that does not fit the grammar is an error rather than a
// wildcard, so ParseRange("latest") fails and Satisfies never admits a version
// against an unparseable range.
//
// Everything is deterministic and depends only on the standard library.
package semver

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ErrInvalidVersion is returned by Parse and the string-based helpers when the
// input is not a valid semantic version.
var ErrInvalidVersion = errors.New("semver: invalid version")

const (
	// semMaxLength bounds the accepted length of a version string, mirroring
	// node-semver's MAX_LENGTH.
	semMaxLength = 256
	// semMaxSafeInteger is the largest value a numeric field may take, mirroring
	// JavaScript's Number.MAX_SAFE_INTEGER used by node-semver.
	semMaxSafeInteger uint64 = 9007199254740991
	// semMaxSafeComponentLength is the maximum digit length coercion will accept
	// for a single numeric component, mirroring MAX_SAFE_COMPONENT_LENGTH.
	semMaxSafeComponentLength = 16
)

// Version is a parsed semantic version. The zero Version is 0.0.0 with no
// prerelease or build identifiers.
type Version struct {
	// Major, Minor and Patch are the three numeric release fields.
	Major uint64
	Minor uint64
	Patch uint64
	// Prerelease holds the dot-separated identifiers following a '-', in order
	// (empty when the version has no prerelease).
	Prerelease []string
	// Build holds the dot-separated build-metadata identifiers following a '+',
	// in order (empty when the version has no build metadata). Build metadata
	// does not affect precedence.
	Build []string
}

// Parse parses a semantic version string, tolerating surrounding whitespace and
// a single optional leading lowercase "v". It returns ErrInvalidVersion
// (wrapped) when the string is not a valid version.
//
// This is node-semver's strict grammar, whose regexp is anchored "^v?" after a
// trim: an uppercase "V" and a leading "=" belong to its loose mode and are
// rejected here. Clean, which mirrors upstream's clean(), strips those prefixes
// before parsing, and a leading "=" inside a range expression is the equality
// operator, handled by the range parser.
func Parse(s string) (*Version, error) {
	if len(s) > semMaxLength {
		return nil, fmt.Errorf("%w: too long", ErrInvalidVersion)
	}
	raw := strings.TrimSpace(s)
	if len(raw) > 0 && raw[0] == 'v' {
		raw = raw[1:]
	}
	if raw == "" {
		return nil, fmt.Errorf("%w: empty", ErrInvalidVersion)
	}

	var build []string
	if i := strings.IndexByte(raw, '+'); i >= 0 {
		b := raw[i+1:]
		raw = raw[:i]
		if b == "" {
			return nil, fmt.Errorf("%w: empty build metadata", ErrInvalidVersion)
		}
		build = strings.Split(b, ".")
		for _, id := range build {
			if !semIsBuildID(id) {
				return nil, fmt.Errorf("%w: build id %q", ErrInvalidVersion, id)
			}
		}
	}

	var pre []string
	if i := strings.IndexByte(raw, '-'); i >= 0 {
		p := raw[i+1:]
		raw = raw[:i]
		if p == "" {
			return nil, fmt.Errorf("%w: empty prerelease", ErrInvalidVersion)
		}
		pre = strings.Split(p, ".")
		for _, id := range pre {
			if !semIsPrereleaseID(id) {
				return nil, fmt.Errorf("%w: prerelease id %q", ErrInvalidVersion, id)
			}
		}
	}

	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: need major.minor.patch", ErrInvalidVersion)
	}
	nums := make([]uint64, 3)
	for i, p := range parts {
		if !semIsNumericID(p) || (len(p) > 1 && p[0] == '0') {
			return nil, fmt.Errorf("%w: numeric field %q", ErrInvalidVersion, p)
		}
		n, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidVersion, err)
		}
		if n > semMaxSafeInteger {
			return nil, fmt.Errorf("%w: numeric field %q too big", ErrInvalidVersion, p)
		}
		nums[i] = n
	}
	return &Version{Major: nums[0], Minor: nums[1], Patch: nums[2], Prerelease: pre, Build: build}, nil
}

// MustParse is like Parse but panics if the version is invalid. It is intended
// for package-level variables and tests where the input is known good.
func MustParse(s string) *Version {
	v, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return v
}

// Valid reports whether s is a well-formed semantic version.
func Valid(s string) bool {
	_, err := Parse(s)
	return err == nil
}

// Clean returns the canonical string form of s when s is a valid version, or ""
// when it is not. Like node-semver's clean() it is more permissive than Parse:
// it first trims surrounding whitespace and strips any run of leading "=" and
// "v" characters, so "=v1.2.3" cleans to "1.2.3".
func Clean(s string) string {
	v, err := Parse(strings.TrimLeft(strings.TrimSpace(s), "=v"))
	if err != nil {
		return ""
	}
	return v.String()
}

// String returns the canonical string representation of the version, including
// any prerelease and build-metadata identifiers.
func (v *Version) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Prerelease) > 0 {
		b.WriteByte('-')
		b.WriteString(strings.Join(v.Prerelease, "."))
	}
	if len(v.Build) > 0 {
		b.WriteByte('+')
		b.WriteString(strings.Join(v.Build, "."))
	}
	return b.String()
}

// Core returns a copy of the version with prerelease and build metadata
// stripped, i.e. just Major.Minor.Patch.
func (v *Version) Core() *Version {
	return &Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch}
}

// Compare returns -1, 0 or +1 as v is less than, equal to, or greater than o
// in semantic-version precedence. Build metadata is ignored.
func (v *Version) Compare(o *Version) int {
	if c := semCmpUint(v.Major, o.Major); c != 0 {
		return c
	}
	if c := semCmpUint(v.Minor, o.Minor); c != 0 {
		return c
	}
	if c := semCmpUint(v.Patch, o.Patch); c != 0 {
		return c
	}
	return semComparePrerelease(v.Prerelease, o.Prerelease)
}

// LessThan reports whether v has lower precedence than o.
func (v *Version) LessThan(o *Version) bool { return v.Compare(o) < 0 }

// GreaterThan reports whether v has higher precedence than o.
func (v *Version) GreaterThan(o *Version) bool { return v.Compare(o) > 0 }

// Equal reports whether v and o have equal precedence (build metadata ignored).
func (v *Version) Equal(o *Version) bool { return v.Compare(o) == 0 }

// IncMajor returns a new version with Major incremented and the lower fields,
// prerelease and build metadata cleared. A pre-major version (X.0.0-pre) is
// promoted to X.0.0 without incrementing Major, matching npm semver.
func (v *Version) IncMajor() *Version {
	major := v.Major
	if v.Minor != 0 || v.Patch != 0 || len(v.Prerelease) == 0 {
		major++
	}
	return &Version{Major: major}
}

// IncMinor returns a new version with Minor incremented, Patch cleared and
// prerelease and build metadata cleared. A pre-minor version (X.Y.0-pre) is
// promoted to X.Y.0 without incrementing Minor, matching npm semver.
func (v *Version) IncMinor() *Version {
	minor := v.Minor
	if v.Patch != 0 || len(v.Prerelease) == 0 {
		minor++
	}
	return &Version{Major: v.Major, Minor: minor}
}

// IncPatch returns a new version with Patch incremented and prerelease and
// build metadata cleared. A prerelease version is promoted to its release form
// without incrementing Patch, matching npm semver.
func (v *Version) IncPatch() *Version {
	if len(v.Prerelease) > 0 {
		return &Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch}
	}
	return &Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
}

// Compare parses two version strings and returns -1, 0 or +1 as a is less than,
// equal to, or greater than b. It panics if either string is invalid; use Valid
// to check first when the inputs are untrusted.
func Compare(a, b string) int {
	return MustParse(a).Compare(MustParse(b))
}

// GT reports whether version string a has higher precedence than b.
func GT(a, b string) bool { return Compare(a, b) > 0 }

// GTE reports whether version string a has precedence greater than or equal to b.
func GTE(a, b string) bool { return Compare(a, b) >= 0 }

// LT reports whether version string a has lower precedence than b.
func LT(a, b string) bool { return Compare(a, b) < 0 }

// LTE reports whether version string a has precedence less than or equal to b.
func LTE(a, b string) bool { return Compare(a, b) <= 0 }

// EQ reports whether version strings a and b have equal precedence.
func EQ(a, b string) bool { return Compare(a, b) == 0 }

// NEQ reports whether version strings a and b differ in precedence.
func NEQ(a, b string) bool { return Compare(a, b) != 0 }

// Sort sorts a slice of version strings ascending, in place, keeping each
// string exactly as it was given. Invalid strings cause a panic; validate
// untrusted input first.
//
// Ties in precedence are broken by build metadata, as npm semver's sort() does
// (it orders with compareBuild, not compare), so "1.0.0+a" sorts before
// "1.0.0+b" and a version with no build metadata sorts before one that has it.
func Sort(versions []string) {
	type entry struct {
		raw string
		v   *Version
	}
	parsed := make([]entry, len(versions))
	for i, s := range versions {
		parsed[i] = entry{raw: s, v: MustParse(s)}
	}
	sort.SliceStable(parsed, func(i, j int) bool {
		return semCompareBuild(parsed[i].v, parsed[j].v) < 0
	})
	for i, e := range parsed {
		versions[i] = e.raw
	}
}

// semCompareBuild is Compare with build metadata as a final tiebreak, mirroring
// node-semver's compareBuild. Build identifiers compare exactly like prerelease
// identifiers, except that having no build metadata sorts first.
func semCompareBuild(a, b *Version) int {
	if c := a.Compare(b); c != 0 {
		return c
	}
	switch {
	case len(a.Build) == 0 && len(b.Build) == 0:
		return 0
	case len(a.Build) == 0:
		return -1
	case len(b.Build) == 0:
		return 1
	}
	n := len(a.Build)
	if len(b.Build) < n {
		n = len(b.Build)
	}
	for i := 0; i < n; i++ {
		if c := semCompareIdentifier(a.Build[i], b.Build[i]); c != 0 {
			return c
		}
	}
	return semCmpInt(len(a.Build), len(b.Build))
}

// Major returns the major field of a version string.
func Major(s string) (uint64, error) {
	v, err := Parse(s)
	if err != nil {
		return 0, err
	}
	return v.Major, nil
}

// Minor returns the minor field of a version string.
func Minor(s string) (uint64, error) {
	v, err := Parse(s)
	if err != nil {
		return 0, err
	}
	return v.Minor, nil
}

// Patch returns the patch field of a version string.
func Patch(s string) (uint64, error) {
	v, err := Parse(s)
	if err != nil {
		return 0, err
	}
	return v.Patch, nil
}

// Prerelease returns the dot-joined prerelease identifiers of a version string
// ("" when there are none).
func Prerelease(s string) (string, error) {
	v, err := Parse(s)
	if err != nil {
		return "", err
	}
	return strings.Join(v.Prerelease, "."), nil
}

// Inc returns the version string incremented at the given release level. It
// accepts every level npm semver's inc() accepts: "major", "minor", "patch",
// "premajor", "preminor", "prepatch", "prerelease" and "release". It returns an
// error for an invalid version or an unknown release level.
//
// The pre* levels bump the corresponding numeric field and then start a
// prerelease at ".0"; "prerelease" bumps the trailing numeric prerelease
// identifier of an existing prerelease, or acts like "prepatch" on a release
// version; and "release" strips the prerelease from a prerelease version and
// fails on a version that has none.
func Inc(s, release string) (string, error) {
	return IncWith(s, release, "")
}

// IncWith is Inc with an explicit prerelease identifier, the second argument of
// npm semver's inc(version, release, identifier). With identifier "beta",
// incrementing "1.2.3" as "prepatch" yields "1.2.4-beta.0", and incrementing
// "1.2.4-beta.0" as "prerelease" yields "1.2.4-beta.1". An empty identifier
// reproduces Inc's bare numeric prerelease.
//
// A non-empty identifier must itself be a legal prerelease: a dot-separated list
// of identifiers, each alphanumeric-with-hyphens or a numeric field without a
// leading zero. npm semver checks the identifier against its PRERELEASE pattern
// the same way and returns null for, say, "bad id" or "01".
func IncWith(s, release, identifier string) (string, error) {
	v, err := Parse(s)
	if err != nil {
		return "", err
	}
	if identifier != "" {
		for _, id := range strings.Split(identifier, ".") {
			if !semIsPrereleaseID(id) {
				return "", fmt.Errorf("%w: invalid prerelease identifier %q", ErrInvalidVersion, identifier)
			}
		}
	}
	switch release {
	case "major":
		return v.IncMajor().String(), nil
	case "minor":
		return v.IncMinor().String(), nil
	case "patch":
		return v.IncPatch().String(), nil
	case "premajor":
		// npm semver clears the prerelease and bumps major unconditionally
		// here. It deliberately does NOT reuse the "major" rule, which skips
		// the bump when minor, patch and prerelease say the version is already
		// an unreleased major (so inc("1.0.0-1","major") is "1.0.0" but
		// inc("1.0.0-1","premajor") is "2.0.0-0").
		base := &Version{Major: v.Major + 1}
		return incPre(base, identifier).String(), nil
	case "preminor":
		// As above: minor is bumped unconditionally after dropping any
		// prerelease, so inc("1.2.0-1","preminor") is "1.3.0-0", not
		// "1.2.0-0".
		base := &Version{Major: v.Major, Minor: v.Minor + 1}
		return incPre(base, identifier).String(), nil
	case "prepatch":
		// Any existing prerelease is irrelevant at this point, so drop it
		// before bumping the patch, matching npm semver.
		base := &Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch}
		return incPre(base.IncPatch(), identifier).String(), nil
	case "prerelease":
		next := v
		if len(v.Prerelease) == 0 {
			next = v.IncPatch()
		} else {
			next = &Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch,
				Prerelease: append([]string(nil), v.Prerelease...)}
		}
		return incPre(next, identifier).String(), nil
	case "release":
		if len(v.Prerelease) == 0 {
			return "", fmt.Errorf("semver: version %q is not a prerelease", s)
		}
		return (&Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch}).String(), nil
	default:
		return "", fmt.Errorf("semver: unknown release level %q", release)
	}
}

// incPre advances v's prerelease, reproducing npm semver's internal "pre"
// increment. With no existing prerelease it starts at "0". Otherwise it
// increments the last numeric identifier, or appends "0" when none of the
// identifiers is numeric. A non-empty identifier replaces the prerelease with
// [identifier, 0] unless the prerelease already starts with that identifier and
// carries a numeric counter, in which case only the counter advances.
func incPre(v *Version, identifier string) *Version {
	pre := append([]string(nil), v.Prerelease...)

	if len(pre) == 0 {
		pre = []string{"0"}
	} else {
		bumped := false
		for i := len(pre) - 1; i >= 0; i-- {
			if n, ok := semParseNumericID(pre[i]); ok {
				pre[i] = strconv.FormatUint(n+1, 10)
				bumped = true
				break
			}
		}
		if !bumped {
			pre = append(pre, "0")
		}
	}

	if identifier != "" {
		if len(v.Prerelease) > 0 && v.Prerelease[0] == identifier && len(v.Prerelease) > 1 {
			if _, ok := semParseNumericID(v.Prerelease[1]); ok {
				// The counter was already advanced above; keep it.
				return &Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch, Prerelease: pre}
			}
		}
		pre = []string{identifier, "0"}
	}

	return &Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch, Prerelease: pre}
}

// Coerce extracts the first version-like sequence from arbitrary text and
// returns it as a canonical version string. Missing minor or patch fields
// default to zero, so "v2" becomes "2.0.0" and "1.2.x build" becomes "1.2.0".
// It returns ErrInvalidVersion when no numeric component is found.
func Coerce(s string) (string, error) {
	// Mirror node-semver's COERCE regex (default: no rtl, no includePrerelease):
	//   (^|[^\d])(\d{1,16})(?:\.(\d{1,16}))?(?:\.(\d{1,16}))?(?:$|[^\d])
	// i.e. find the leftmost run of at most three dot-separated numeric
	// components, each 1..16 digits, where the whole match is bounded by
	// non-digits (or the string ends). A digit run longer than 16 digits is not
	// matched and is skipped over entirely.
	n := len(s)
	i := 0
	for i < n {
		if !semIsDigitByte(s[i]) {
			i++
			continue
		}
		// i is the start of a maximal digit run (i == 0 or s[i-1] is non-digit).
		j := i
		for j < n && semIsDigitByte(s[j]) {
			j++
		}
		if j-i > semMaxSafeComponentLength {
			// too-long component: skip the entire run and keep scanning.
			i = j
			continue
		}
		major := s[i:j]
		minor, patch := "", ""
		k := j
		if k < n && s[k] == '.' && k+1 < n && semIsDigitByte(s[k+1]) {
			m := k + 1
			for m < n && semIsDigitByte(s[m]) {
				m++
			}
			if m-(k+1) <= semMaxSafeComponentLength {
				minor = s[k+1 : m]
				k = m
				if k < n && s[k] == '.' && k+1 < n && semIsDigitByte(s[k+1]) {
					p := k + 1
					for p < n && semIsDigitByte(s[p]) {
						p++
					}
					if p-(k+1) <= semMaxSafeComponentLength {
						patch = s[k+1 : p]
					}
				}
			}
		}
		v := &Version{}
		var err error
		if v.Major, err = semCoerceField(major); err != nil {
			return "", err
		}
		if v.Minor, err = semCoerceField(minor); err != nil {
			return "", err
		}
		if v.Patch, err = semCoerceField(patch); err != nil {
			return "", err
		}
		return v.String(), nil
	}
	return "", fmt.Errorf("%w: no numeric component", ErrInvalidVersion)
}

func semIsDigitByte(c byte) bool { return c >= '0' && c <= '9' }

// semCoerceField parses a coerced numeric component (empty means 0). It rejects
// values above MAX_SAFE_INTEGER and components with a leading zero: upstream
// coerce() reassembles the matched digit runs verbatim and hands the result to
// parse(), so "01.02.03" coerces to nothing rather than to 1.2.3.
func semCoerceField(s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, fmt.Errorf("%w: numeric field %q has a leading zero", ErrInvalidVersion, s)
	}
	nv, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidVersion, err)
	}
	if nv > semMaxSafeInteger {
		return 0, fmt.Errorf("%w: numeric field %q too big", ErrInvalidVersion, s)
	}
	return nv, nil
}

// --- comparison helpers -----------------------------------------------------

func semCmpUint(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func semComparePrerelease(a, b []string) int {
	// A version without prerelease outranks one with a prerelease.
	switch {
	case len(a) == 0 && len(b) == 0:
		return 0
	case len(a) == 0:
		return 1
	case len(b) == 0:
		return -1
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if c := semCompareIdentifier(a[i], b[i]); c != 0 {
			return c
		}
	}
	return semCmpInt(len(a), len(b))
}

func semCompareIdentifier(a, b string) int {
	an, aok := semParseNumericID(a)
	bn, bok := semParseNumericID(b)
	switch {
	case aok && bok:
		return semCmpUint(an, bn)
	case aok:
		// numeric identifiers have lower precedence than alphanumeric
		return -1
	case bok:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

func semCmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func semParseNumericID(s string) (uint64, bool) {
	if !semIsNumericID(s) {
		return 0, false
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func semIsNumericID(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func semIsPrereleaseID(s string) bool {
	if s == "" {
		return false
	}
	if semIsNumericID(s) {
		return len(s) == 1 || s[0] != '0' // no leading zeros in numeric prerelease
	}
	return semIsBuildID(s)
}

func semIsBuildID(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '-') {
			return false
		}
	}
	return true
}
