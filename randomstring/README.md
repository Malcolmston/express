# randomstring — standard-library port of the npm "randomstring" library

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/randomstring.svg)](https://pkg.go.dev/github.com/malcolmston/express/randomstring)

Package randomstring is a standard-library port of the npm "randomstring"
library. It generates random strings from named character-set presets or a
custom character set using crypto/rand for unbiased sampling.

Random strings are a common building block for tokens, temporary passwords,
session identifiers, file names, and test fixtures. This package packages that
need behind a small API: pick a length and either a named charset or a literal
set of characters, and receive a string of that length drawn from those
characters. Because the source of randomness is crypto/rand, the output is
suitable for security-sensitive identifiers, unlike helpers built on
math/rand.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/randomstring"
```

## Usage

This is the package's own `ExampleGenerate`, so it compiles and its output is
asserted on every `go test ./randomstring/`.

```go
s, err := randomstring.Generate(16, "numeric")
	if err != nil {
		panic(err)
	}
	allDigits := strings.Trim(s, "0123456789") == ""
	fmt.Println(len(s), allDigits)
```

```
16 true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Generate(length int, charset string) (string, error)` | Generate returns a random string of the requested length using a preset character set. |
| `func GenerateFrom(length int, chars string) (string, error)` | GenerateFrom returns a random string of the requested length drawn uniformly from the literal character set chars using crypto/rand. |
| `func GenerateWith(opts Options) (string, error)` | GenerateWith returns a random string built from opts, reproducing the character-set pipeline of the npm original's generate(options): the charset… |
| `func New() (string, error)` | New returns a 32-character alphanumeric random string. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options mirrors the options object accepted by the npm original's randomstring.generate(options). |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/randomstring).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
