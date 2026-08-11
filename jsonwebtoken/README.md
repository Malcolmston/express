# jsonwebtoken — signs and verifies JSON Web Tokens (JWTs) using HMAC (HS256/HS384/HS512)

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/jsonwebtoken.svg)](https://pkg.go.dev/github.com/malcolmston/express/jsonwebtoken)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/jsonwebtoken)

Package jsonwebtoken signs and verifies JSON Web Tokens (JWTs) using HMAC
(HS256/HS384/HS512). It is a Go port of the widely used npm "jsonwebtoken"
package, covering the symmetric-key signing and verification path that most
applications reach for, implemented entirely with the standard library
(crypto/hmac, crypto/sha256, crypto/sha512, crypto/subtle and encoding/json).

JWTs are the de facto token format for stateless authentication and for
passing signed claims between services: a server signs a small JSON payload
with a shared secret, hands the resulting string to a client, and can later
trust that string because tampering would invalidate the signature. This port
exists so Go programs can produce and check those tokens with the same mental
model as the Node library — Sign to mint a token, Verify to validate one,
and Decode to peek at the payload — without depending on a third-party
module.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/jsonwebtoken"
```

## Usage

This is the package's own `ExampleDecode`, so it compiles and its output is
asserted on every `go test ./jsonwebtoken/`.

```go
secret := []byte("my-fixed-secret")
	token, _ := jsonwebtoken.Sign(jsonwebtoken.Claims{
		"sub": "user-42",
	}, secret, nil)

	claims, err := jsonwebtoken.Decode(token)
	if err != nil {
		fmt.Println("decode error:", err)
		return
	}
	fmt.Println(claims["sub"])
```

```
user-42
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Sign(claims Claims, secret []byte, opts *SignOptions) (string, error)` | Sign builds and signs a JWT. |

### Types

| Type | What it is |
| --- | --- |
| `Claims` | Claims represents the JWT payload. |
| `Complete` | Complete is the fully decomposed form of a token, the equivalent of Node's jwt.decode(token, {complete: true}). |
| `SignOptions` | SignOptions configures signing. |
| `VerifyOptions` | VerifyOptions constrains what Verify will accept. |

<details>
<summary><code>Claims</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Decode(token string) (Claims, error)` | Decode returns the payload claims without verifying the signature. |
| `func Verify(token string, secret []byte, opts ...*VerifyOptions) (Claims, error)` | Verify checks the signature and standard time claims, returning the claims. |

</details>

<details>
<summary><code>Complete</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func DecodeComplete(token string) (*Complete, error)` | DecodeComplete returns the token's header, payload and signature WITHOUT verifying anything, mirroring jwt.decode(token, {complete: true}). |

</details>

### Variables

`ErrInvalidSignature`, `ErrTokenExpired`, `ErrTokenNotValidYet`, `ErrInvalidToken`, `ErrUnsupportedAlg`, `ErrInvalidAudience`, `ErrInvalidIssuer`, `ErrInvalidSubject`, `ErrInvalidJwtID`, `ErrTokenTooOld`, `ErrMissingIssuedAt`, `ErrMissingKey`, `ErrAsymmetricKey`, `ErrClaimConflict`, `ErrInvalidClaimType`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/jsonwebtoken).

## Measured parity

Compared case-for-case against **`jsonwebtoken@9.0.3`** by the harness in
[`parity/express/nested/jsonwebtoken`](https://github.com/malcolmston/go/tree/main/parity/express/nested/jsonwebtoken):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 129 |
| Matching | 120 |
| Mismatching | 0 |
| Declared deviations | 9 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/jsonwebtoken/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
