# ms — converts between time durations and human readable strings

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/ms.svg)](https://pkg.go.dev/github.com/malcolmston/express/ms)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/ms)

Package ms converts between time durations and human readable strings, modeled
on the npm "ms" package. It parses strings such as "2h", "1d" or "2.5 days"
into time.Durations and formats time.Durations back into short ("2h") or long
("2 hours") forms. The npm original is one of the most widely depended-on
utilities in the Node ecosystem, used throughout Express and its middleware
for things like cookie max-ages, cache lifetimes and timeout configuration;
this package is a standard-library-only Go port of it.

The point of the package is to let humans and configuration files express
durations in a friendly notation while your code works with a real
time.Duration. You write "30d" or "2.5 hours" in a config value or a
command-line flag and Parse turns it into something you can add to a time.Time
or hand to a timer. Going the other way, Format and FormatLong turn a computed
duration back into a compact label for logs, UIs or error messages. All
functions are pure, allocation-light and safe for concurrent use.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/ms"
```

## Usage

This is the package's own `ExampleFormat`, so it compiles and its output is
asserted on every `go test ./ms/`.

```go
fmt.Println(ms.Format(2 * time.Hour))
	fmt.Println(ms.Format(500 * time.Millisecond))
	fmt.Println(ms.Format(-3 * time.Second))
```

```
2h
500ms
-3s
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Format(d time.Duration) string` | Format returns the short human readable form of d, choosing the largest unit whose magnitude is at least one: years ("y"), months ("mo"), weeks… |
| `func FormatLong(d time.Duration) string` | FormatLong returns the long human readable form of d, such as "2 hours" or "1 minute", with correct pluralization. |
| `func Parse(s string) (time.Duration, error)` | Parse converts a human readable duration string such as "2h", "1d", "-1.5h", "100" or "2.5 days" into a time.Duration. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/ms).

## Measured parity

Compared case-for-case against **`ms@4.0.0-nightly.202508271359`** by the harness in
[`parity/express/nested/ms`](https://github.com/malcolmston/go/tree/main/parity/express/nested/ms):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 180 |
| Matching | 178 |
| Mismatching | 0 |
| Declared deviations | 2 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/ms/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
