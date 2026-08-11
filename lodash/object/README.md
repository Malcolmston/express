# object — ports the "Object" category of the npm lodash library to Go

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/lodash/object.svg)](https://pkg.go.dev/github.com/malcolmston/express/lodash/object)

Package object ports the "Object" category of the npm lodash library to Go,
providing the map- and structure-oriented helpers such as Keys, Values,
Entries, Pick, Omit, MapKeys, MapValues, Invert, Assign, Merge, Defaults, Get,
Set, Has, Unset, Update, Clone, CloneDeep, IsEqual and Transform. In
JavaScript these operate on plain objects; here they operate on Go maps and,
for the dynamically-typed path and merge helpers, on the any-typed tree of
map`string`any and []any values that a decoded JSON document produces. The
package depends only on the Go standard library.

Reach for this package when you are manipulating decoded JSON or other
loosely-typed nested data and want the ergonomic, batteries-included
vocabulary that lodash offers on the front end: extracting a subset of keys,
re-keying or re-mapping values, deeply merging configuration layers, reading
or writing a value several levels down by a single "a.b.c" path, or comparing
two trees for structural equality. The statically-typed helpers (Keys, Values,
Entries, Pick, Omit, MapKeys, MapValues, Invert, FindKey and friends) are
written with Go generics so they work over any comparable key type and any
value type without reflection or boxing.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/lodash/object"
```

## Usage

This is the package's own `ExampleFindKey`, so it compiles and its output is
asserted on every `go test ./object/`.

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
	key, ok := object.FindKey(m, func(v int, k string) bool { return v == 2 })
	fmt.Println(key, ok)
```

```
b true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Assign(dst map[string]any, sources ...map[string]any) map[string]any` | Assign copies the own enumerable keys of each source into dst, with later sources overwriting earlier ones. |
| `func AssignWith[K comparable, V any](dst, src map[K]V, customizer func(dstVal, srcVal V, key K) V) map[K]V` | AssignWith merges src into a copy of dst, resolving each key with customizer, which receives the destination and source values (the destination value… |
| `func At[K comparable, V any](m map[K]V, keys ...K) []V` | At returns the values of m at the given keys, in the order the keys are supplied. |
| `func Clone(value any) any` | Clone returns a shallow copy of value. |
| `func CloneDeep(value any) any` | CloneDeep returns a deep copy of value, recursively copying nested map`string`any and []any containers so that mutating the result never affects the… |
| `func Defaults(dst map[string]any, sources ...map[string]any) map[string]any` | Defaults assigns, for each source in sources, the values of keys that are missing (or hold the zero value equivalent of "undefined") in dst. |
| `func DefaultsDeep(dst map[string]any, sources ...map[string]any) map[string]any` | DefaultsDeep is like Defaults except it recurses into nested map`string`any values, filling in missing keys at every depth. |
| `func EveryEntry[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) bool` | EveryEntry reports whether predicate returns true for every entry of m. |
| `func FindEntry[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) (K, V, bool)` | FindEntry returns an arbitrary entry of m for which predicate is true, along with true; if no entry matches it returns zero values and false. |
| `func FindKey[V any](m map[string]V, predicate func(value V, key string) bool) (string, bool)` | FindKey returns the key of the first entry of m for which predicate returns true, along with true. |
| `func ForIn(m map[string]any, iteratee func(value any, key string) bool)` | ForIn iterates over the entries of m, invoking iteratee for each. |
| `func ForOwn(m map[string]any, iteratee func(value any, key string) bool)` | ForOwn iterates over the own enumerable entries of m, invoking iteratee for each. |
| `func FromEntries[K comparable, V any](pairs []Pair[K, V]) map[K]V` | FromEntries builds a map from a slice of Pair. |
| `func Get(root any, path string, defaultValue ...any) any` | Get retrieves the value at the given dot-notation path within root. |
| `func Has(root any, path string) bool` | Has reports whether root contains a value at the given dot-notation path. |
| `func Invert[K comparable, V comparable](m map[K]V) map[V]K` | Invert returns a new map composed of the inverted keys and values of m. |
| `func InvertBy[K comparable, V any, R comparable](m map[K]V, iteratee func(value V) R) map[R][]K` | InvertBy returns a new map composed of keys generated from the results of running each value of m through iteratee. |
| `func IsEqual(a, b any) bool` | IsEqual reports whether a and b are deeply, structurally equal. |
| `func Keys[K comparable, V any](m map[K]V) []K` | Keys returns the keys of m. |
| `func MapKeys[K comparable, V any, R comparable](m map[K]V, iteratee func(value V, key K) R) map[R]V` | MapKeys returns a new map with the same values as m but with keys produced by running each entry through iteratee. |
| `func MapValues[K comparable, V any, R any](m map[K]V, iteratee func(value V, key K) R) map[K]R` | MapValues returns a new map with the same keys as m but with values produced by running each entry through iteratee. |
| `func Merge(dst map[string]any, sources ...map[string]any) map[string]any` | Merge recursively merges the sources into dst. |
| `func Omit[K comparable, V any](m map[K]V, keys ...K) map[K]V` | Omit returns a new map with the given keys removed from m. |
| `func OmitBy[K comparable, V any](m map[K]V, predicate func(value V, key K) bool) map[K]V` | OmitBy returns a new map with the entries of m for which predicate returns true removed. |
| `func Pick[K comparable, V any](m map[K]V, keys ...K) map[K]V` | Pick returns a new map composed of the given keys picked from m. |
| `func PickBy[K comparable, V any](m map[K]V, predicate func(value V, key K) bool) map[K]V` | PickBy returns a new map composed of the entries of m for which predicate returns true. |
| `func ReduceEntries[K comparable, V any, R any](m map[K]V, iteratee func(acc R, key K, value V) R, accumulator R) R` | ReduceEntries folds the entries of m into an accumulator using iteratee. |
| `func Set(root any, path string, value any) any` | Set sets the value at the given dot-notation path within root, creating intermediate map`string`any containers as needed. |
| `func SomeEntry[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) bool` | SomeEntry reports whether predicate returns true for at least one entry of m. |
| `func Transform(m map[string]any, iteratee func(acc map[string]any, value any, key string) bool, accumulator map[string]any) map[string]any` | Transform is a variant of Reduce for objects. |
| `func Unset(root any, path string) bool` | Unset removes the value at the given dot-notation path from root. |
| `func Update(root any, path string, updater func(current any) any) any` | Update is like Set except the new value is produced by running the current value at path (or nil if absent) through updater. |
| `func Values[K comparable, V any](m map[K]V) []V` | Values returns the values of m. |

### Types

| Type | What it is |
| --- | --- |
| `Pair` | Pair is a single key/value entry, as produced by Entries. |

<details>
<summary><code>Pair</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Entries[K comparable, V any](m map[K]V) []Pair[K, V]` | Entries returns the key/value pairs of m as a slice of Pair. |
| `func ToPairs[K comparable, V any](m map[K]V) []Pair[K, V]` | ToPairs is an alias for Entries, matching lodash's naming. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/lodash/object).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
