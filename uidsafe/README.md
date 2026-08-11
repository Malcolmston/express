# uidsafe — generates cryptographically secure, URL-safe unique identifiers

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/uidsafe.svg)](https://pkg.go.dev/github.com/malcolmston/express/uidsafe)

Package uidsafe generates cryptographically secure, URL-safe unique
identifiers. It is a standard-library-only Go port of the npm "uid-safe"
library, the token generator used by express-session and cookie-session to
mint session ids and by other middleware that needs an unguessable, opaque
string. The name captures the two guarantees it provides: the identifier is
safe to embed unescaped in URLs, cookies, and headers, and it is drawn from a
cryptographically secure source so it cannot be predicted by an attacker.

The generation strategy is deliberately simple. Bytes reads n bytes from
crypto/rand, the operating system's CSPRNG, and encodes them with base64
URL-safe encoding (the alphabet that uses '-' and '_' in place of '+' and '/')
without any '=' padding. Because base64 packs three input bytes into four
output characters, the resulting string is longer than n and its length is not
equal to n; for example 18 random bytes produce a 24-character token. Callers
who care about the exact string length should choose n accordingly rather than
assuming a one-to-one mapping.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/uidsafe"
```

## Usage

This is the package's own `ExampleBytes`, so it compiles and its output is
asserted on every `go test ./uidsafe/`.

```go
s, err := uidsafe.Bytes(18)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(s))
```

```
24
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Bytes(n int) (string, error)` | Bytes returns a URL-safe base64 string generated from n cryptographically random bytes. |
| `func MustBytes(n int) string` | MustBytes is like Bytes but panics if n is negative or the random source fails. |
| `func New() (string, error)` | New returns a token built from DefaultBytes random bytes, the size express-session uses for a session id. |

### Constants

`DefaultBytes`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/uidsafe).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
