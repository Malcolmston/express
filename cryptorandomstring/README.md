# cryptorandomstring — standard-library port of the npm "crypto-random-string" library

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/cryptorandomstring.svg)](https://pkg.go.dev/github.com/malcolmston/express/cryptorandomstring)

Package cryptorandomstring is a standard-library port of the npm
"crypto-random-string" library. It generates cryptographically strong random
strings via crypto/rand with unbiased sampling, and offers the same set of
named character-set presets as the original module.

Use this package anywhere you need a random token that must be hard to guess:
session identifiers, API keys, password-reset tokens, one-time codes,
filenames for temporary uploads, and similar. Because it draws from
crypto/rand rather than math/rand, the output is suitable for security
contexts where predictability would be a vulnerability. When you only need a
random hex string, the Hex convenience wrapper covers the common case; for
anything else, Generate takes an Options struct that selects a preset
character set or supplies a completely custom one.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/cryptorandomstring"
```

## Usage

This is the package's own `ExampleGenerate`, so it compiles and its output is
asserted on every `go test ./cryptorandomstring/`.

```go
s, err := cryptorandomstring.Generate(cryptorandomstring.Options{
		Length: 8,
		Type:   "numeric",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	allDigits := true
	for _, r := range s {
		if r < '0' || r > '9' {
			allDigits = false
		}
	}
	fmt.Println(len(s), allDigits)
```

```
8 true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Generate(opts Options) (string, error)` | Generate returns a secure random string according to opts. |
| `func Hex(length int) (string, error)` | Hex returns a secure random hexadecimal string of the given length. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options controls the generated string. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/cryptorandomstring).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
