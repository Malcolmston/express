# sqlstring — provides MySQL-style value and identifier escaping and query formatting

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/sqlstring.svg)](https://pkg.go.dev/github.com/malcolmston/express/sqlstring)

Package sqlstring provides MySQL-style value and identifier escaping and query
formatting. It is a Go port of the npm "sqlstring" library (the escaping
engine used by the popular mysqljs/mysql driver), reimplemented with only the
Go standard library. Its purpose is to let callers build MySQL statements from
dynamic data without opening the door to SQL injection.

The package is used whenever an application composes SQL text that embeds
user-supplied or otherwise untrusted values and cannot (or does not wish to)
rely solely on the database driver's own parameter binding. Rather than
concatenating raw values into a query, callers pass those values through
Escape (for data) or EscapeID (for identifiers), or let Format do both at once
while filling placeholders. The escaped output is safe to splice into a query
string because every value is turned into a self-contained, correctly quoted
SQL literal.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/sqlstring"
```

## Usage

This is the package's own `ExampleEscape`, so it compiles and its output is
asserted on every `go test ./sqlstring/`.

```go
fmt.Println(sqlstring.Escape(42))
	fmt.Println(sqlstring.Escape(true))
	fmt.Println(sqlstring.Escape(nil))
	fmt.Println(sqlstring.Escape("bob"))
	fmt.Println(sqlstring.Escape("x' OR '1'='1"))
```

```
42
true
NULL
'bob'
'x\' OR \'1\'=\'1'
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Escape(val any) string` | Escape returns a safely-quoted SQL literal for val suitable for MySQL. |
| `func EscapeID(id string) string` | EscapeID wraps an identifier in backticks, doubling any embedded backticks so the result is a single safe MySQL identifier. |
| `func EscapeIDForbidQualified(id string) string` | EscapeIDForbidQualified wraps an identifier in backticks, treating a dot as an ordinary character rather than a qualifier separator: "table.col"… |
| `func EscapeIDList(ids []string, forbidQualified bool) string` | EscapeIDList escapes each identifier and joins the results with ", ", mirroring the upstream behavior of escapeId when handed an array. |
| `func Format(sql string, args []any) string` | Format substitutes placeholders in sql left-to-right with values from args. |
| `func FormatArgs(sql string, args ...any) string` | FormatArgs is the variadic form of Format, matching the way the upstream SqlString.format is normally called from JavaScript. |

### Types

| Type | What it is |
| --- | --- |
| `Raw` | Raw is a fragment of SQL that Escape and Format must splice in literally rather than quote. |
| `SQLStringer` | SQLStringer is implemented by values that render themselves as raw, already-safe SQL. |

<details>
<summary><code>Raw</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func (r Raw) ToSqlString() string` | ToSqlString returns the raw SQL text, satisfying SQLStringer. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/sqlstring).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
