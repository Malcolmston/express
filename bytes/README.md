# bytes — converts between byte counts and human readable strings

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/bytes.svg)](https://pkg.go.dev/github.com/malcolmston/express/bytes)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/bytes)

Package bytes converts between byte counts and human readable strings, modeled
on the npm "bytes" package by TJ Holowaychuk. Units are binary (1KB = 1024
bytes). It reimplements that package's two headline operations, parsing a size
string into a number of bytes and formatting a number of bytes into a size
string, using only the Go standard library.

This is the utility Express and its middleware use to accept and display byte
sizes in a form people actually write. You use Parse to turn a configuration
value or request field such as "100kb" or "1.5mb" into an int64 you can
compare against, for example when enforcing a request body size limit, and you
use Format (or FormatOpts) to turn a raw count back into a compact label such
as "1.5MB" for logs, dashboards, or API responses. Keeping both directions in
one place means the same unit conventions apply whether a size is being read
in or printed out.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/bytes"
```

## Usage

This is the package's own `ExampleParse`, so it compiles and its output is
asserted on every `go test ./bytes/`.

```go
n, err := bytes.Parse("1.5MB")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(n)
```

```
1572864
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Format(n int64) string` | Format renders n as a human readable string using the largest fitting binary unit, e.g. |
| `func FormatOpts(n int64, opts FormatOptions) string` | FormatOpts renders n as a human readable string using the supplied options. |
| `func Parse(s string) (int64, error)` | Parse converts a human readable byte string such as "1KB" or "1.5MB" into an int64 count of bytes. |

### Types

| Type | What it is |
| --- | --- |
| `FormatOptions` | FormatOptions controls how Format renders a byte count. |

### Constants

`B`, `KB`, `MB`, `GB`, `TB`, `PB`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/bytes).

## Measured parity

Compared case-for-case against **`bytes@3.1.2`** by the harness in
[`parity/express/nested/bytes`](https://github.com/malcolmston/go/tree/main/parity/express/nested/bytes):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 230 |
| Matching | 218 |
| Mismatching | 0 |
| Declared deviations | 12 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/bytes/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
