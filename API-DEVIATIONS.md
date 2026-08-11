# API deviations

Every package in this repository is a port of an existing library, and the
cross-language parity harness in the aggregator
(`parity/express/`, `parity/express/nested/<pkg>/`) scores each one against its
real upstream. A handful of differences are deliberate: Go's type system, its
error convention, or a bounded integer type makes exact parity either impossible
or a worse API. Those are recorded here, one section per package, and the
corresponding parity case is tagged `"deviation"` so the harness counts it
separately from a bug.

Nothing may be listed here merely because it is inconvenient to fix. If upstream
is right and the port is wrong, the port gets fixed.

## `ms` (npm `ms`, oracle `ms@4.0.0-nightly.202508271359`)

Parity: 178/178 compared cases (100%), 2 deviations.
Harness: `parity/express/nested/ms/`.

| deviation | parity case | why |
| --- | --- | --- |
| `Parse` returns a range error for a duration beyond about ±292 years, where upstream returns an ordinary (unbounded) JavaScript number. | `p-err-overflow-1e30`, `p-err-overflow-years` | The port's return type is `time.Duration`, an `int64` count of nanoseconds. `ms.parse("100000y")` has no representable answer. Go leaves an out-of-range `float64 → int64` conversion implementation-defined, so the alternative to an error is a silently wrapped duration — the worse of the two behaviours. |
| There is no single dual-dispatch entry point equivalent to upstream's default `ms(value, options?)`, which switches on `typeof value`. | `x-ms-string-*`, `x-ms-number-*`, `p-err-ms-empty` | Runtime dispatch on argument type is not a thing a statically typed API should emulate. The two directions are `Parse` and `Format`/`FormatLong`; both arms are covered by cases and behave identically to upstream. Likewise `parseStrict`, which upstream distinguishes from `parse` only by a TypeScript template-literal argument type, maps onto the same `Parse`. |
| The `{ long: true }` format option is a separate function, `FormatLong`, rather than an options argument. | every `fl-*` case | Same reason: one boolean option does not earn an options struct in Go, and the split is checkable at compile time. |

## `bytes` (npm `bytes`, oracle `bytes@3.1.2`)

Parity: 218/218 compared cases (100%), 12 deviations.
Harness: `parity/express/nested/bytes/`.

| deviation | parity case | why |
| --- | --- | --- |
| `Parse` returns an overflow error where upstream returns an oversized JavaScript number: past `int64` (`"1000000000000000000000"`, `"99999999999999999999pb"`) and at the very top of the range, where the decimal literal rounds up to the double 2^63 (`"9223372036854775807"`). | `pb-1e21`, `pb-1e23`, `pb-pb-overflow`, `pb-int64max` | The port's return type is `int64`. Go leaves an out-of-range `float64 → int64` conversion implementation-defined, so the alternative to an error is a silently wrapped byte count — and this value is routinely used as a request-body size limit, where a wrapped number is a security bug, not a rounding artefact. |
| `Parse` has no overload for a number argument; upstream's `bytes.parse(1024)` returns its argument unchanged. | `pn-int`, `pn-zero`, `pn-neg`, `pn-frac` | Go has no function overloading, and a `Parse(string)` that also accepted `any` to implement an identity function would be a worse API. `pn-frac` is doubly unrepresentable: upstream returns the fraction unfloored. |
| `FormatOpts` clamps `DecimalPlaces` into 0..100 instead of failing; `Number.prototype.toFixed` throws a `RangeError` outside that range. | `fo-dp-negative`, `fo-dp-101` | `FormatOpts` returns a bare `string` and has no error channel, and a formatter is not worth panicking from. Clamping is the only remaining option that still produces the documented output for every in-range value. |
| An explicit `decimalPlaces: null` keeps the port's default of 2, where upstream treats it as `toFixed(0)`. | `fo-dp-null` | Upstream checks only `options.decimalPlaces !== undefined`, so JavaScript's `null` slips through and `ToIntegerOrInfinity(null)` is 0. A Go `*int` has exactly one "absent" state; JSON `null` and an omitted key both decode to `nil`. |
| Formatting a count above 2^53 prints the exact `int64`, where upstream can only print the nearest double. | `fo-unit-b-2pow53` | `bytes.format` takes a double, so 2^53+1 has already become 2^53 before the function is entered. The port is strictly more accurate here and there is nothing to reproduce. |

Everything else about this port is upstream's behaviour, quirks included, and
several of those quirks are load-bearing surprises rather than bugs in the port:
`Parse` falls back to JavaScript's `parseInt(value, 10)` for any string the unit
grammar rejects (so `"1zb"`, `"1024 please"` and `" 1kb"` are 1, 1024 and **1
byte** rather than errors), `Format` breaks a rounding tie away from zero like
`toFixed` rather than to even like `strconv.FormatFloat`, and its trailing-zero
trimmer is upstream's regexp, which cannot touch a fractional part that begins
with a zero (`"1.010KB"` stays).
