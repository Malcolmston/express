# util — ports helpers from lodash's "Util" category (cond, stubArray, stubTrue/False/String

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/lodash/util.svg)](https://pkg.go.dev/github.com/malcolmston/express/lodash/util)

Package util ports helpers from lodash's "Util" category (cond, stubArray,
stubTrue/False/String, toPath and a stepped range) to idiomatic, generic Go.
These are the small building blocks lodash exposes for control flow and for
providing constant "stub" values, offered here without any third-party
dependency.

Cond builds a function from an ordered list of predicate/transform pairs and
runs the transform of the first matching predicate, the Go analogue of
lodash.cond and a compact alternative to a switch when the branches are data.
The Stub* helpers return fresh empty values (StubArray and StubObject each
allocate a new, non-nil container) and the constant StubTrue/StubFalse/
StubString mirror lodash's stubs for use as default callbacks. ToPath parses a
lodash-style property path such as "a.b[0].c" into its string segments, and
RangeStep generates an arithmetic sequence with an explicit step, filling the
gap left by the plain Range helpers.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/lodash/util"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Cond[T, R any](pairs ...CondPair[T, R]) func(T) (R, bool)` | Cond returns a function that evaluates each pair's predicate in order and, for the first whose predicate is satisfied, returns that pair's transform… |
| `func RangeStep(start, end, step int) []int` | RangeStep returns the arithmetic sequence starting at start, advancing by step, up to but not including end. |
| `func StubArray[T any]() []T` | StubArray returns a new empty, non-nil slice of T, mirroring lodash.stubArray. |
| `func StubFalse() bool` | StubFalse always returns false, mirroring lodash.stubFalse. |
| `func StubObject[K comparable, V any]() map[K]V` | StubObject returns a new empty, non-nil map, mirroring lodash.stubObject. |
| `func StubString() string` | StubString always returns the empty string, mirroring lodash.stubString. |
| `func StubTrue() bool` | StubTrue always returns true, mirroring lodash.stubTrue. |
| `func ToIndex(segment string) (int, bool)` | ToIndex parses a single path segment as a non-negative array index, returning the index and true when the segment is a valid unsigned integer literal. |
| `func ToPath(path string) []string` | ToPath splits a lodash-style property path into its segments. |

### Types

| Type | What it is |
| --- | --- |
| `CondPair` | CondPair is a single predicate/transform branch for Cond: When selects the branch and Then produces its result. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/lodash/util).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
