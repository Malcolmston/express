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
// a string into a *Version; MustParse panics instead of returning an error and
// is convenient for constants and tests. Valid reports whether a string is a
// well-formed version, and Clean normalizes surrounding whitespace and an
// optional leading "v" or "=".
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
// x-ranges ("1.x", "1.2.*", "*"), hyphen ranges ("1.2.3 - 2.3.4"), space- or
// comma-separated AND terms within a set, and "||" separating alternative sets.
// Prerelease versions only satisfy a range when a comparator in the same tuple
// names the same Major.Minor.Patch with its own prerelease, matching npm's
// default (includePrerelease=false) behaviour.
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

// Parse parses a semantic version string, tolerating an optional leading "v" or
// "=" and surrounding whitespace. It returns ErrInvalidVersion (wrapped) when
// the string is not a valid version.
func Parse(s string) (*Version, error) {
	if len(s) > semMaxLength {
		return nil, fmt.Errorf("%w: too long", ErrInvalidVersion)
	}
	raw := strings.TrimSpace(s)
	raw = strings.TrimPrefix(raw, "=")
	raw = strings.TrimSpace(raw)
	if len(raw) > 0 && (raw[0] == 'v' || raw[0] == 'V') {
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

// Clean returns the canonical string form of s (dropping a leading "v"/"=" and
// surrounding whitespace) when s is a valid version, or "" when it is not.
func Clean(s string) string {
	v, err := Parse(s)
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

// Sort sorts a slice of version strings ascending, in place. Invalid strings
// cause a panic; validate untrusted input first.
func Sort(versions []string) {
	parsed := make([]*Version, len(versions))
	for i, s := range versions {
		parsed[i] = MustParse(s)
	}
	sort.SliceStable(parsed, func(i, j int) bool { return parsed[i].Compare(parsed[j]) < 0 })
	for i, v := range parsed {
		versions[i] = v.String()
	}
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

// Inc returns the version string incremented at the given release level, which
// must be "major", "minor" or "patch". It returns an error for an invalid
// version or an unknown release level.
func Inc(s, release string) (string, error) {
	v, err := Parse(s)
	if err != nil {
		return "", err
	}
	switch release {
	case "major":
		return v.IncMajor().String(), nil
	case "minor":
		return v.IncMinor().String(), nil
	case "patch":
		return v.IncPatch().String(), nil
	default:
		return "", fmt.Errorf("semver: unknown release level %q", release)
	}
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

// semCoerceField parses a coerced numeric component (empty means 0) and rejects
// values above MAX_SAFE_INTEGER, mirroring parse() failing on an out-of-range
// component.
func semCoerceField(s string) (uint64, error) {
	if s == "" {
		return 0, nil
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
