# jwtdecode — decodes the payload of a JSON Web Token without verifying its signature

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/jwtdecode.svg)](https://pkg.go.dev/github.com/malcolmston/express/jwtdecode)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/jwtdecode)

Package jwtdecode decodes the payload of a JSON Web Token without verifying
its signature. It is a Go port of the npm "jwt-decode" package, implemented
with only the standard library (encoding/base64 and encoding/json).

The jwt-decode library is a tiny, browser-friendly helper whose single job is
to turn a JWT string into a plain object so an application can read the claims
it carries — a user id, a display name, an expiry timestamp — for display
or routing decisions. It exists precisely because full JWT libraries are
heavier than needed when all you want is to look inside a token you already
hold. This port serves the same purpose in Go: it exposes Decode for the
payload and DecodeHeader for the header, each returning a
map`string`interface{} shaped exactly like the decoded JSON.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/jwtdecode"
```

## Usage

This is the package's own `ExampleDecode`, so it compiles and its output is
asserted on every `go test ./jwtdecode/`.

```go
claims, err := jwtdecode.Decode(token)
	if err != nil {
		panic(err)
	}
	fmt.Println(claims["sub"])
	fmt.Println(claims["name"])
	fmt.Printf("%.0f\n", claims["iat"])
```

```
1234567890
John Doe
1516239022
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Decode(token string) (map[string]interface{}, error)` | Decode decodes the payload (second segment) of a JWT without verifying it. |
| `func DecodeHeader(token string) (map[string]interface{}, error)` | DecodeHeader decodes the header (first segment) of a JWT without verifying it. |

### Variables

`ErrInvalidToken`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/jwtdecode).

## Measured parity

Compared case-for-case against **`jwt-decode@4.0.0`** by the harness in
[`parity/express/nested/jwtdecode`](https://github.com/malcolmston/go/tree/main/parity/express/nested/jwtdecode):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 49 |
| Matching | 45 |
| Mismatching | 0 |
| Declared deviations | 4 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/jwtdecode/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
