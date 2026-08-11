# base32 — encodes and decodes byte data using the RFC 4648 base32 alphabet

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/base32.svg)](https://pkg.go.dev/github.com/malcolmston/express/base32)

Package base32 encodes and decodes byte data using the RFC 4648 base32
alphabet, providing a small convenience wrapper over the standard library's
encoding/base32. It plays the role of the npm base32 helpers used in the wider
Express port, offering the handful of encode/decode operations those libraries
expose while delegating the actual bit-packing to Go's battle-tested
implementation.

Base32 represents arbitrary binary data as text drawn from a 32-character
alphabet, so it is useful whenever data must survive systems that are
case-insensitive, that mangle mixed case, or that only tolerate a limited set
of letters and digits. Common uses include human-transcribable identifiers,
TOTP/2FA secret keys, filenames on case-folding filesystems, and any token
that might be spoken aloud or typed by hand. It trades compactness for
robustness: base32 output is longer than base64, but its alphabet avoids the
case sensitivity and punctuation that make base64 fragile in those settings.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/base32"
```

## Usage

This is the package's own `ExampleDecode`, so it compiles and its output is
asserted on every `go test ./base32/`.

```go
data, err := base32.Decode("nbuq")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%s\n", data)
```

```
hi
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Decode(s string) ([]byte, error)` | Decode decodes a base32 string, case-insensitively and tolerating missing padding. |
| `func Encode(data []byte) string` | Encode encodes data using the uppercase RFC 4648 alphabet with padding. |
| `func EncodeNoPadding(data []byte) string` | EncodeNoPadding encodes data without the "=" padding characters. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/base32).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
