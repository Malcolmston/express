# mimetypes — provides MIME type lookups by filename or extension and reverse lookups from MIME type

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/mimetypes.svg)](https://pkg.go.dev/github.com/malcolmston/express/mimetypes)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/mimetypes)

Package mimetypes provides MIME type lookups by filename or extension and
reverse lookups from MIME type to extension. It is a port of the npm
mime-types package, the module Express uses behind res.type/res.contentType
and that content-negotiation middleware relies on, implemented using only the
Go standard library. It answers the everyday questions "what Content-Type
should I send for this file?" and, going the other way, "what file extension
does this MIME type correspond to?".

The package is useful anywhere you serve files, set response headers, or map
uploads to a canonical type. Rather than depending on the operating system's
mime.types database (whose contents vary by machine), it ships a curated,
self-contained table covering the common web, image, video, audio, font,
document and archive types, so results are identical across platforms. The
table is a subset of mime-db, but every entry in it is taken *from* mime-db
rather than chosen by hand, so a type the table knows resolves exactly as
upstream resolves it. All functions are pure lookups over that table, do no
I/O, and are safe for concurrent use.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/mimetypes"
```

## Usage

This is the package's own `ExampleLookup_unknown`, so it compiles and its output is
asserted on every `go test ./mimetypes/`.

```go
t, ok := mimetypes.Lookup("archive.unknownext")
	fmt.Printf("%q %v\n", t, ok)
```

```
"" false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Charset(mimeType string) (string, bool)` | Charset returns the charset for a MIME type, or false if none applies. |
| `func ContentType(typeOrExt string) (string, bool)` | ContentType returns a full Content-Type value for a MIME type or extension. |
| `func Extension(mimeType string) (string, bool)` | Extension returns the first file extension (without a leading dot) that mime-db lists for a MIME type, ignoring any parameters. |
| `func Lookup(pathOrExt string) (string, bool)` | Lookup returns the MIME type for a filename, path, or extension. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/mimetypes).

## Measured parity

Compared case-for-case against **`mime-types@2.1.35`** by the harness in
[`parity/express/nested/mimetypes`](https://github.com/malcolmston/go/tree/main/parity/express/nested/mimetypes):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 188 |
| Matching | 188 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/mimetypes/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
