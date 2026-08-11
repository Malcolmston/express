# shortid — standard-library-only short, URL-friendly unique id generator inspired by the npm

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/shortid.svg)](https://pkg.go.dev/github.com/malcolmston/express/shortid)

Package shortid is a standard-library-only short, URL-friendly unique id
generator inspired by the npm "shortid" library. It produces compact,
non-sequential-looking strings suitable for use in URLs, filenames, and other
contexts where a full UUID would be unnecessarily long. Every character comes
from a 64-character URL-safe alphabet, so ids never require percent-encoding.

An id is built from two concatenated components. The first is a time
component: the current time in Unix milliseconds is written out in the
alphabet's base (64), least-significant digit first, which yields a handful of
characters that advance as the clock advances. The second is a random
component of six characters, each drawn independently from crypto/rand via
math/big, which disambiguates ids created within the same millisecond. The two
parts together are typically 7 to 14 characters long depending on the
magnitude of the current timestamp.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/shortid"
```

## Usage

This is the package's own `ExampleIsValid`, so it compiles and its output is
asserted on every `go test ./shortid/`.

```go
fmt.Println(shortid.IsValid("abc-_XYZ"))
	fmt.Println(shortid.IsValid(""))
	fmt.Println(shortid.IsValid("has space"))
```

```
true
false
false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Generate() (string, error)` | Generate returns a new short id combining a time component and a random component. |
| `func IsValid(s string) bool` | IsValid reports whether s is a plausible shortid: it must be at least six characters long and contain only characters from the current alphabet. |
| `func SetAlphabet(a string) error` | SetAlphabet replaces the alphabet used by Generate and IsValid. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/shortid).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
