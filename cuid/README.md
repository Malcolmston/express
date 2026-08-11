# cuid — generates collision-resistant unique identifiers

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/cuid.svg)](https://pkg.go.dev/github.com/malcolmston/express/cuid)

Package cuid generates collision-resistant unique identifiers, a Go port of
the npm "cuid" package (in the style of its cuid2 successor). Ids are
URL-safe, hard to guess, and safe to generate across many hosts and processes
without any central coordination or shared state.

Reach for a cuid when you need a unique identifier for a database row, an
event, an uploaded object, or a distributed message, and you want something
friendlier than a raw UUID: the output is a short lowercase alphanumeric
string that always begins with a letter, so it is safe to use in URLs, as an
HTML element id, or as a filename without escaping. Unlike a sequential
integer key, a cuid reveals little about how many ids have been issued and
does not require a round trip to a central allocator, which makes it
convenient for client-side generation and for sharded systems.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/cuid"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./cuid/`.

```go
id, err := cuid.New()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(len(id), cuid.IsCuid(id))
```

```
24 true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func IsCuid(s string) bool` | IsCuid reports whether s is a syntactically valid cuid2. |
| `func IsSlug(s string) bool` | IsSlug reports whether s has the length of a cuid slug. |
| `func New() (string, error)` | New returns a cuid2 with the default length. |
| `func NewLength(n int) (string, error)` | NewLength returns a cuid2 of the given length (must be between 2 and 32). |
| `func Slug() (string, error)` | Slug returns a short, slug-style identifier, a port of the upstream cuid.slug() helper. |

### Constants

`DefaultLength`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/cuid).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
