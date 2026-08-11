# base58 — standard-library-only Go implementation of Base58 and Base58Check encoding

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/base58.svg)](https://pkg.go.dev/github.com/malcolmston/express/base58)

Package base58 is a standard-library-only Go implementation of Base58 and
Base58Check encoding, the compact, ambiguity-free binary-to-text encoding
popularised by Bitcoin and used across the npm ecosystem (bs58, base58check,
bs58check) for keys, addresses and short identifiers. Base58 uses the digits
and letters excluding the visually confusable 0, O, I and l, so encoded
strings survive hand transcription.

Encode and Decode implement the raw Bitcoin-alphabet encoding, preserving
leading zero bytes as leading '1' characters. CheckEncode and CheckDecode
implement Base58Check: they prepend a one-byte version, append a four-byte
checksum (the first four bytes of the double SHA-256 of the version-and-
payload), and verify that checksum on decode, returning ErrInvalidChecksum
when it does not match. Decode returns ErrInvalidCharacter for any character
outside the alphabet.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/base58"
```

## Usage

This is the package's own `ExampleCheckEncode`, so it compiles and its output is
asserted on every `go test ./base58/`.

```go
s := base58.CheckEncode([]byte{0xde, 0xad, 0xbe, 0xef}, 0x00)

	version, payload, err := base58.CheckDecode(s)
	fmt.Printf("%#x %x %v\n", version, payload, err)

	corrupt := []byte(s)
	if corrupt[1] == 'z' {
		corrupt[1] = 'y'
	} else {
		corrupt[1] = 'z'
	}
	_, _, err = base58.CheckDecode(string(corrupt))
	fmt.Println(errors.Is(err, base58.ErrInvalidChecksum))

	_, _, err = base58.CheckDecode("2g")
	fmt.Println(errors.Is(err, base58.ErrTooShort))
```

```
0x0 deadbeef <nil>
true
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func CheckDecode(s string) (version byte, payload []byte, err error)` | CheckDecode decodes a Base58Check string, verifies its four-byte checksum and returns the version byte and the payload. |
| `func CheckEncode(payload []byte, version byte) string` | CheckEncode returns the Base58Check encoding of payload with the given version byte: it encodes version\|\|payload\|\|checksum, where checksum is the… |
| `func Decode(s string) ([]byte, error)` | Decode returns the bytes represented by the Base58-encoded string s. |
| `func Encode(input []byte) string` | Encode returns the Base58 encoding of input. |

### Variables

`ErrInvalidCharacter`, `ErrInvalidChecksum`, `ErrTooShort`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/base58).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
