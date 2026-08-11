# base62 — standard-library-only Go implementation of Base62 encoding, the alphanumeric (0-9, A-Z

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/base62.svg)](https://pkg.go.dev/github.com/malcolmston/express/base62)

Package base62 is a standard-library-only Go implementation of Base62
encoding, the alphanumeric (0-9, A-Z, a-z) binary-to-text scheme used by URL
shorteners and short-id generators across the npm ecosystem (base62, base-x).
Because the alphabet contains only characters that are safe in URLs and
identifiers, Base62 is a common choice for compact, human-friendly tokens.

Encode and Decode convert between arbitrary byte slices and Base62 strings,
preserving leading zero bytes as leading '0' characters. EncodeInt and
DecodeInt convert unsigned 64-bit integers, the form most short-id schemes
use. Decode and DecodeInt return ErrInvalidCharacter for input outside the
alphabet. The implementation uses math/big for the base conversion, is
deterministic, and depends only on the standard library.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/base62"
```

## Usage

This is the package's own `ExampleDecodeInt`, so it compiles and its output is
asserted on every `go test ./base62/`.

```go
n, err := base62.DecodeInt("4gfFC4")
	fmt.Println(n, err)

	_, err = base62.DecodeInt("abc-def")
	fmt.Println(errors.Is(err, base62.ErrInvalidCharacter))

	_, err = base62.DecodeInt("zzzzzzzzzzzzzzzzzzzz")
	fmt.Println(errors.Is(err, base62.ErrOverflow))
```

```
4294967296 <nil>
true
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Decode(s string) ([]byte, error)` | Decode returns the bytes represented by the Base62-encoded string s. |
| `func DecodeInt(s string) (uint64, error)` | DecodeInt returns the unsigned integer represented by the Base62 string s. |
| `func Encode(input []byte) string` | Encode returns the Base62 encoding of input. |
| `func EncodeInt(n uint64) string` | EncodeInt returns the Base62 encoding of an unsigned integer. |

### Constants

`Alphabet`

### Variables

`ErrInvalidCharacter`, `ErrOverflow`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/base62).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
