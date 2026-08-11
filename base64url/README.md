# base64url — encodes and decodes byte data using the RFC 4648 url-safe base64 alphabet without

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/base64url.svg)](https://pkg.go.dev/github.com/malcolmston/express/base64url)

Package base64url encodes and decodes byte data using the RFC 4648 url-safe
base64 alphabet without padding, as used in JWTs and other web tokens. It
ports the npm base64url package's surface on top of the standard library's
encoding/base64, exposing the same encode, decode, and standard-base64
interconversion helpers while delegating the underlying transform to Go.

You want this package whenever base64 data has to travel through a URL, an
HTTP header, a filename, or a JSON web token. Ordinary base64 uses "+" and
"/", which are reserved or unsafe in those positions, and its "=" padding is
often stripped by intermediaries; base64url swaps those two characters for "-"
and "_" and omits padding so the text can be dropped into a query string, a
path segment, or a JWT segment without any further percent-encoding. It is the
encoding you see in JWT headers and payloads, in OAuth tokens, and in many
web-facing identifiers.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/base64url"
```

## Usage

This is the package's own `ExampleDecodeString`, so it compiles and its output is
asserted on every `go test ./base64url/`.

```go
s, err := base64url.DecodeString("PDxoZWxsbz4-")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(s)
```

```
<<hello>>
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Decode(s string) ([]byte, error)` | Decode decodes a url-safe base64 string, tolerating missing padding. |
| `func DecodeString(s string) (string, error)` | DecodeString decodes a url-safe base64 string and returns it as a string. |
| `func Encode(data []byte) string` | Encode encodes data as RFC 4648 url-safe base64 without padding. |
| `func EncodeString(s string) string` | EncodeString encodes a string as url-safe base64 without padding. |
| `func FromBase64(s string) string` | FromBase64 converts a standard base64 string to a base64url string. |
| `func ToBase64(s string) string` | ToBase64 converts a base64url string to a standard base64 string with padding. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/base64url).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
