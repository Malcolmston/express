# totp — implements time-based one-time passwords as defined in RFC 6238

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/totp.svg)](https://pkg.go.dev/github.com/malcolmston/express/totp)

Package totp implements time-based one-time passwords as defined in RFC 6238,
the short numeric codes that authenticator apps such as Google Authenticator,
Authy, and 1Password display for two-factor authentication. It is a
stdlib-only Go port of the TOTP functionality found in the popular npm
libraries "otplib", "speakeasy", and "totp": generating the current code from
a shared secret and verifying a user-supplied code against a small window of
recent and upcoming time steps. It validates against the RFC 6238 Appendix B
test vectors.

TOTP is the time-based cousin of HOTP (RFC 4226). Both derive a code by
applying HMAC to an 8-byte big-endian counter keyed by the shared secret and
then performing the RFC 4226 "dynamic truncation" to extract a 31-bit number
that is reduced modulo 10^digits and zero-padded to a fixed width. The one
difference is what feeds the counter: HOTP uses an explicit counter that both
sides advance on each use, while TOTP computes the counter from the clock as
floor(unixTime / period). Because the server and the authenticator app share
only the secret and read the same wall clock, they independently arrive at the
same code with no network round trip.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/totp"
```

## Usage

This is the package's own `ExampleGenerateAt`, so it compiles and its output is
asserted on every `go test ./totp/`.

```go
secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	code, err := totp.GenerateAt(secret, time.Unix(59, 0).UTC(), &totp.Options{Digits: 8})
	if err != nil {
		panic(err)
	}
	fmt.Println(code)
```

```
94287082
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Counter(t time.Time, opts *Options) (uint64, error)` | Counter returns the RFC 6238 time-step counter T = (unixTime - T0) / X for t. |
| `func DecodeSecret(base32Secret string) ([]byte, error)` | DecodeSecret decodes a Base32 shared secret using the same lenient rules Generate applies: the input is upper-cased, spaces and dashes are removed,… |
| `func Generate(base32Secret string, opts *Options) (string, error)` | Generate returns the TOTP value for the current time. |
| `func GenerateAt(base32Secret string, t time.Time, opts *Options) (string, error)` | GenerateAt returns the TOTP value for the given time. |
| `func GenerateWithCounter(base32Secret string, counter uint64, opts *Options) (string, error)` | GenerateWithCounter returns the RFC 4226 HOTP value for an explicit counter, the primitive TOTP is built on. |
| `func TimeRemaining(opts *Options) time.Duration` | TimeRemaining returns how long the code for the current time stays valid. |
| `func TimeRemainingAt(t time.Time, opts *Options) time.Duration` | TimeRemainingAt returns how long the code for t stays valid: the time left in t's step. |
| `func Verify(base32Secret, code string, opts *Options, window int) bool` | Verify reports whether code is valid within +/- window steps of the current time. |
| `func VerifyAt(base32Secret, code string, t time.Time, opts *Options, window int) (delta int, ok bool)` | VerifyAt reports whether code is valid within +/- window steps of t, and returns the matching step offset relative to t's own step. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures TOTP generation and verification. |

### Constants

`MaxDigits`

### Variables

`ErrInvalidSecret`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/totp).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
