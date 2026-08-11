# cookiesignature — signs and verifies cookie values using HMAC-SHA256

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/cookiesignature.svg)](https://pkg.go.dev/github.com/malcolmston/express/cookiesignature)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/cookiesignature)

Package cookiesignature signs and verifies cookie values using HMAC-SHA256,
mirroring the behavior of the npm cookie-signature library. It is the
mechanism Express and cookie-parser use to produce "signed cookies": a
tamper-evident wrapper around an otherwise plaintext cookie value that lets a
server detect whether a value it later receives was the one it originally
issued.

Reach for this package whenever a value is round-tripped through an untrusted
party (a browser, a proxy, a query string) but must not be forgeable. Signing
does not hide or encrypt the value; anyone can read it. What signing
guarantees is integrity and authenticity: only a holder of the secret can
produce a signature that Unsign will accept, so a client cannot alter the
value (for example, escalate a session id or a role flag) without invalidating
the signature. Typical uses are session identifiers, CSRF tokens, and
remember-me cookies.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/cookiesignature"
```

## Usage

This is the package's own `ExampleUnsign`, so it compiles and its output is
asserted on every `go test ./cookiesignature/`.

```go
signed := cookiesignature.Sign("hello", "secret")
	value, ok := cookiesignature.Unsign(signed, "secret")
	fmt.Println(value, ok)

	_, ok = cookiesignature.Unsign(signed, "wrong-secret")
	fmt.Println(ok)
```

```
hello true
false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Sign(value, secret string) string` | Sign returns value with an HMAC-SHA256 signature appended, in the form "value.signature". |
| `func Unsign(signed, secret string) (string, bool)` | Unsign verifies a signed value produced by Sign against secret. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/cookiesignature).

## Measured parity

Compared case-for-case against **`cookie-signature@1.2.2`** by the harness in
[`parity/express/nested/cookiesignature`](https://github.com/malcolmston/go/tree/main/parity/express/nested/cookiesignature):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 54 |
| Matching | 54 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/cookiesignature/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
