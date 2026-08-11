# hotp — implements HMAC-based one-time passwords as defined in RFC 4226

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/hotp.svg)](https://pkg.go.dev/github.com/malcolmston/express/hotp)

Package hotp implements HMAC-based one-time passwords as defined in RFC 4226,
used for counter-based two-factor authentication codes. It is a
standard-library port of the npm "hotp" / "otplib" style HOTP primitive,
exposing the two operations applications actually need: Generate, which
derives a one-time code from a shared secret and a counter, and Verify, which
checks a user-supplied code against the expected value.

HOTP was standardized by the IETF in RFC 4226 (2005) as the foundation of the
Initiative for Open Authentication (OATH). It exists so that a server and a
hardware token or mobile app, sharing only a secret key, can independently
arrive at the same short numeric code without any network round trip. HOTP is
the counter-based cousin of TOTP: where TOTP feeds the current time window
into the algorithm, HOTP feeds a monotonically increasing counter that both
sides advance on each successful use. This port targets the same behavior as
the JavaScript reference implementations and validates against the RFC 4226
Appendix D test vectors.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/hotp"
```

## Usage

This is the package's own `ExampleVerify`, so it compiles and its output is
asserted on every `go test ./hotp/`.

```go
secret := []byte("12345678901234567890")
	fmt.Println(hotp.Verify(secret, 0, "755224", 6))
	fmt.Println(hotp.Verify(secret, 0, "000000", 6))
```

```
true
false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Generate(secret []byte, counter uint64, digits int) string` | Generate computes an RFC 4226 HOTP value for the given secret and counter, returning a zero-padded decimal string of the requested number of digits. |
| `func Verify(secret []byte, counter uint64, code string, digits int) bool` | Verify reports whether code matches the HOTP value for the secret and counter. |
| `func VerifyWindow(secret []byte, counter uint64, code string, digits, window int) (uint64, bool)` | VerifyWindow checks code against every counter in [counter, counter+window] and reports the counter that matched. |

### Constants

`MaxDigits`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/hotp).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
