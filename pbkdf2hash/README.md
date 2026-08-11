# pbkdf2hash — implements PBKDF2-HMAC-SHA256 key derivation together with a self-describing

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/pbkdf2hash.svg)](https://pkg.go.dev/github.com/malcolmston/express/pbkdf2hash)

Package pbkdf2hash implements PBKDF2-HMAC-SHA256 key derivation together with
a self-describing, textual password-hash encoding. It occupies the same niche
as the npm pbkdf2-password module and the broader family of PHC/@node-rs-style
password hashers: it turns a plaintext password into a storable string and can
later verify a candidate password against that string without keeping the
original around. Everything here is built on the Go standard library
(crypto/hmac, crypto/sha256, crypto/rand and crypto/subtle), so it has no
third-party dependencies.

You reach for this package whenever you need to persist passwords safely.
Storing plaintext (or a plain SHA-256) is unsafe because a leaked database
immediately exposes every credential. PBKDF2 defends against that by making
each guess deliberately expensive: an attacker must repeat many thousands of
HMAC evaluations per candidate password, which throttles offline brute-force
and dictionary attacks. The per-hash random salt additionally defeats rainbow
tables and ensures two users with the same password produce different hashes.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/pbkdf2hash"
```

## Usage

This is the package's own `ExampleHash`, so it compiles and its output is
asserted on every `go test ./pbkdf2hash/`.

```go
encoded, err := pbkdf2hash.Hash("s3cret", 1000)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(pbkdf2hash.Verify("s3cret", encoded))
	fmt.Println(pbkdf2hash.Verify("wrong", encoded))
```

```
true
false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func DeriveKey(password, salt []byte, iterations, keyLen int) []byte` | DeriveKey derives a key of keyLen bytes from password and salt using PBKDF2 with HMAC-SHA256 over the given number of iterations. |
| `func Hash(password string, iterations int) (string, error)` | Hash derives a 32-byte key from password with a random 16-byte salt and returns an encoded string: pbkdf2_sha256$<iterations>$<hexsalt>$<hexkey>. |
| `func HashWith(password string, opts Options) (string, error)` | HashWith derives a key from password using opts and returns the encoded string. |
| `func NeedsRehash(encoded string, iterations int) bool` | NeedsRehash reports whether an encoded hash was produced with fewer than the given iteration count, which is the signal to re-hash the password at… |
| `func Params(encoded string) (algorithm string, iterations, saltLen, keyLen int, err error)` | Params reports the algorithm tag, iteration count, salt length in bytes and key length in bytes recorded in an encoded hash. |
| `func Verify(password, encoded string) bool` | Verify parses an encoded hash, re-derives the key from password, and compares in constant time. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures HashWith. |

### Constants

`DefaultIterations`, `DefaultSaltLength`, `DefaultKeyLength`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/pbkdf2hash).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
