# jwtauth — provides Express middleware that verifies HS256-signed JSON Web Tokens supplied via

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/jwtauth.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/jwtauth)

Package jwtauth provides Express middleware that verifies HS256-signed JSON
Web Tokens supplied via the Authorization header, gating access to protected
routes. It fills the same role as the Node "express-jwt" middleware (commonly
paired with "jsonwebtoken"), but is a deliberately small, stdlib-only
reimplementation: HMAC-SHA256 verification is done with crypto/hmac and
crypto/sha256, and JSON decoding with encoding/json, so it pulls in no
third-party dependencies.

Use this middleware to protect API endpoints with stateless bearer-token
authentication. A client obtains an HS256 JWT elsewhere (a login endpoint, an
identity provider) and presents it on each request as "Authorization: Bearer
<token>"; the middleware verifies the signature and exposes the decoded claims
to downstream handlers. It is appropriate when the signer and verifier share a
symmetric secret. It does not mint tokens — only verify them — and it only
supports the HS256 algorithm, so RS256/ES256 or asymmetric-key setups are out
of scope.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/jwtauth"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that verifies an HS256 JWT from the "Authorization: Bearer <token>" header. |

### Types

| Type | What it is |
| --- | --- |
| `Claims` | Claims represents the decoded JWT payload. |
| `Options` | Options configures the JWT middleware. |

<details>
<summary><code>Claims</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Verify(token string, secret []byte) (Claims, error)` | Verify validates an HS256 JWT against secret and returns its claims. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/jwtauth).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
