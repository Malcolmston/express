# otpauth — builds and parses otpauth:// key URIs

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/otpauth.svg)](https://pkg.go.dev/github.com/malcolmston/express/otpauth)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/otpauth)

Package otpauth builds and parses otpauth:// key URIs, a stdlib-only Go port
of the URI portion of the npm "otpauth" library
(https://www.npmjs.com/package/otpauth). One-time passwords are the short
numeric codes used for two-factor authentication; the otpauth:// URI is the
de-facto standard, originally defined by Google Authenticator, for handing a
shared secret and its parameters to an authenticator app, most commonly by
encoding the URI in a QR code that the user scans during account setup.

This package concerns itself with the transport format — constructing a
correct otpauth:// URI from a Config and parsing one back into a Config. It
covers both one-time-password families defined by the standard: HOTP (HMAC
based one-time passwords, RFC 4226), which are counter-based and advance each
time a code is consumed, and TOTP (time-based one-time passwords, RFC 6238),
which derive the code from the current time divided into fixed-length steps.
The Type field selects between them, defaulting to "totp" when left empty.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/otpauth"
```

## Usage

This is the package's own `ExampleConfig_Generate`, so it compiles and its output is
asserted on every `go test ./otpauth/`.

```go
c := otpauth.Config{
		Secret:    "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
		Digits:    8,
		Period:    30,
		Algorithm: "SHA1",
	}
	code, err := c.Generate(time.Unix(59, 0))
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
| `func DecodeSecret(secret string) ([]byte, error)` | DecodeSecret decodes a base32 shared secret such as "JBSWY3DPEHPK3PXP" into its raw bytes. |
| `func NormalizeSecret(secret string) string` | NormalizeSecret canonicalises a base32 secret the way the npm library's Secret does: whitespace removed, upper-cased and '=' padding stripped. |
| `func URL(c Config) string` | URL builds an otpauth:// URI from the config. |

### Types

| Type | What it is |
| --- | --- |
| `Config` | Config describes an otpauth key URI. |

<details>
<summary><code>Config</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Parse(rawurl string) (Config, error)` | Parse parses an otpauth:// URI into a Config. |
| `func (c Config) CounterAt(t time.Time) (uint64, error)` | CounterAt returns the counter value the config uses at time t: the fixed Config.Counter for an HOTP config, or the RFC 6238 time step floor(unix(t) /… |
| `func (c Config) Generate(t time.Time) (string, error)` | Generate returns the one-time password for the config at time t. |
| `func (c Config) GenerateCounter(counter uint64) (string, error)` | GenerateCounter computes the one-time password for an explicit counter value, the RFC 4226 primitive that both HOTP and TOTP are built on. |
| `func (c Config) Validate(code string, t time.Time, window int) (delta int, ok bool)` | Validate checks code against the config and reports the delta, in counter steps, between the code and the expected value. |

</details>

### Constants

`DefaultDigits`, `DefaultPeriod`, `DefaultAlgorithm`, `MaxDigits`

### Variables

`ErrInvalidSecret`, `ErrUnsupportedAlgorithm`, `ErrUnsupportedType`, `ErrInvalidURI`, `ErrMissingSecret`, `ErrMissingCounter`, `ErrInvalidDigits`, `ErrInvalidPeriod`, `ErrInvalidAlgorithm`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/otpauth).

## Measured parity

Compared case-for-case against **`otpauth@9.5.1`** by the harness in
[`parity/express/nested/otpauth`](https://github.com/malcolmston/go/tree/main/parity/express/nested/otpauth):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 164 |
| Matching | 158 |
| Mismatching | 0 |
| Declared deviations | 6 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/otpauth/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
