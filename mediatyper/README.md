# mediatyper — parses and formats HTTP media type strings of the form type/subtype+suffix; params

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/mediatyper.svg)](https://pkg.go.dev/github.com/malcolmston/express/mediatyper)

Package mediatyper parses and formats HTTP media type strings of the form
type/subtype+suffix; params. It is a port of the npm media-typer package (the
small module used internally by Express and the wider connect/body middleware
ecosystem) implemented with nothing beyond the Go standard library. Media
types are the "content" identifiers carried in headers such as Content-Type
and Accept, and this package gives you a structured, round-trippable
representation of them following the grammar in RFC 6838 and the
token/quoted-string rules of RFC 7231/7230.

You reach for this package whenever you need to inspect or rewrite a media
type rather than merely compare it as an opaque string. Typical uses are
splitting "application/vnd.api+json; charset=utf-8" into its top-level type,
its subtype, its structured "+json" suffix and its parameters, or building
such a string back up from those pieces so that it is guaranteed to be well
formed. Because it only manipulates syntax it performs no I/O, keeps no state,
and is safe for concurrent use.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/mediatyper"
```

## Usage

This is the package's own `ExampleFormat`, so it compiles and its output is
asserted on every `go test ./mediatyper/`.

```go
s, err := mediatyper.Format(mediatyper.MediaType{
		Type:       "text",
		Subtype:    "html",
		Parameters: map[string]string{"charset": "utf-8"},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(s)
```

```
text/html; charset=utf-8
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Format(m MediaType) (string, error)` | Format reconstructs a media type string from a MediaType value. |

### Types

| Type | What it is |
| --- | --- |
| `MediaType` | MediaType represents a parsed media type. |

<details>
<summary><code>MediaType</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Parse(s string) (MediaType, error)` | Parse parses a media type string. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/mediatyper).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
