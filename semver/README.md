# semver — standard-library-only Go port of the npm "semver" package

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/semver.svg)](https://pkg.go.dev/github.com/malcolmston/express/semver)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/semver)

Package semver is a standard-library-only Go port of the npm "semver" package
(https://www.npmjs.com/package/semver), the reference implementation of
Semantic Versioning 2.0.0 (https://semver.org) used throughout the Node and
Express ecosystems. It parses, compares, increments, coerces and range-matches
version strings with the same semantics JavaScript projects rely on, so
tooling that reasons about dependency versions can be ported to Go without
pulling in a third-party dependency.

A version is Major.Minor.Patch with optional dot-separated prerelease and
build-metadata identifiers, for example "1.2.3-alpha.1+build.5". Parse turns a
string into a *Version, accepting surrounding whitespace and a single optional
lowercase "v" exactly as node-semver's strict grammar does — an uppercase
"V" or a leading "=" is loose-mode-only and is rejected. MustParse panics
instead of returning an error and is convenient for constants and tests. Valid
reports whether a string is a well-formed version, and Clean is the more
permissive normalizer: it trims whitespace and strips any run of leading "="
and "v" characters before parsing.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/semver"
```

## Usage

This is the package's own `ExampleCoerce`, so it compiles and its output is
asserted on every `go test ./semver/`.

```go
for _, s := range []string{"v1.2", "1", "release-3.4.5-rc"} {
		v, err := semver.Coerce(s)
		fmt.Printf("%-16q -> %v %v\n", s, v, err)
	}
```

```
"v1.2"           -> 1.2.0 <nil>
"1"              -> 1.0.0 <nil>
"release-3.4.5-rc" -> 3.4.5 <nil>
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Clean(s string) string` | Clean returns the canonical string form of s when s is a valid version, or "" when it is not. |
| `func Coerce(s string) (string, error)` | Coerce extracts the first version-like sequence from arbitrary text and returns it as a canonical version string. |
| `func Compare(a, b string) int` | Compare parses two version strings and returns -1, 0 or +1 as a is less than, equal to, or greater than b. |
| `func EQ(a, b string) bool` | EQ reports whether version strings a and b have equal precedence. |
| `func GT(a, b string) bool` | GT reports whether version string a has higher precedence than b. |
| `func GTE(a, b string) bool` | GTE reports whether version string a has precedence greater than or equal to b. |
| `func Inc(s, release string) (string, error)` | Inc returns the version string incremented at the given release level. |
| `func IncWith(s, release, identifier string) (string, error)` | IncWith is Inc with an explicit prerelease identifier, the second argument of npm semver's inc(version, release, identifier). |
| `func LT(a, b string) bool` | LT reports whether version string a has lower precedence than b. |
| `func LTE(a, b string) bool` | LTE reports whether version string a has precedence less than or equal to b. |
| `func Major(s string) (uint64, error)` | Major returns the major field of a version string. |
| `func MaxSatisfying(versions []string, constraint string) (string, bool)` | MaxSatisfying returns the highest version in versions that satisfies the range, or "" (and ok=false) if none do. |
| `func MinSatisfying(versions []string, constraint string) (string, bool)` | MinSatisfying returns the lowest version in versions that satisfies the range, or "" (and ok=false) if none do. |
| `func Minor(s string) (uint64, error)` | Minor returns the minor field of a version string. |
| `func NEQ(a, b string) bool` | NEQ reports whether version strings a and b differ in precedence. |
| `func Patch(s string) (uint64, error)` | Patch returns the patch field of a version string. |
| `func Prerelease(s string) (string, error)` | Prerelease returns the dot-joined prerelease identifiers of a version string ("" when there are none). |
| `func Satisfies(version, constraint string) bool` | Satisfies reports whether the version string satisfies the range string. |
| `func Sort(versions []string)` | Sort sorts a slice of version strings ascending, in place, keeping each string exactly as it was given. |
| `func Valid(s string) bool` | Valid reports whether s is a well-formed semantic version. |

### Types

| Type | What it is |
| --- | --- |
| `Range` | Range is a compiled semantic-version range: a disjunction (OR) of tuples, where each tuple is a conjunction (AND) of comparators. |
| `Version` | Version is a parsed semantic version. |

<details>
<summary><code>Range</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func ParseRange(expr string) (*Range, error)` | ParseRange compiles a range expression (for example "^1.2.0", ">=1.0.0 <2.0.0", "1.x \|\| 2.x", "1.2.3 - 2.3.4"). |
| `func (r *Range) String() string` | String returns the original range expression the Range was parsed from. |
| `func (r *Range) Test(v *Version) bool` | Test reports whether the version satisfies the range. |

</details>

<details>
<summary><code>Version</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func MustParse(s string) *Version` | MustParse is like Parse but panics if the version is invalid. |
| `func Parse(s string) (*Version, error)` | Parse parses a semantic version string, tolerating surrounding whitespace and a single optional leading lowercase "v". |
| `func (v *Version) Compare(o *Version) int` | Compare returns -1, 0 or +1 as v is less than, equal to, or greater than o in semantic-version precedence. |
| `func (v *Version) Core() *Version` | Core returns a copy of the version with prerelease and build metadata stripped, i.e. |
| `func (v *Version) Equal(o *Version) bool` | Equal reports whether v and o have equal precedence (build metadata ignored). |
| `func (v *Version) GreaterThan(o *Version) bool` | GreaterThan reports whether v has higher precedence than o. |
| `func (v *Version) IncMajor() *Version` | IncMajor returns a new version with Major incremented and the lower fields, prerelease and build metadata cleared. |
| `func (v *Version) IncMinor() *Version` | IncMinor returns a new version with Minor incremented, Patch cleared and prerelease and build metadata cleared. |
| `func (v *Version) IncPatch() *Version` | IncPatch returns a new version with Patch incremented and prerelease and build metadata cleared. |
| `func (v *Version) LessThan(o *Version) bool` | LessThan reports whether v has lower precedence than o. |
| `func (v *Version) String() string` | String returns the canonical string representation of the version, including any prerelease and build-metadata identifiers. |

</details>

### Variables

`ErrInvalidVersion`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/semver).

## Measured parity

Compared case-for-case against **`semver@7.8.5`** by the harness in
[`parity/express/nested/semver`](https://github.com/malcolmston/go/tree/main/parity/express/nested/semver):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 320 |
| Matching | 313 |
| Mismatching | 0 |
| Declared deviations | 7 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/semver/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
