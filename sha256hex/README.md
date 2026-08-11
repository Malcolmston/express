# sha256hex — provides small, allocation-light helpers that hash bytes or strings and return

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/sha256hex.svg)](https://pkg.go.dev/github.com/malcolmston/express/sha256hex)

Package sha256hex provides small, allocation-light helpers that hash bytes or
strings and return the digest as a lowercase hexadecimal string. It is a
stdlib-only port of the "sha256 hex" idiom exposed by JavaScript hashing
libraries such as js-sha256 (and its sibling js-sha1 / js-md5 packages), whose
default call form hashes an input and returns its digest already encoded as a
lowercase hex string. The same output is what Node's built-in
crypto.createHash("sha256").update(data).digest("hex") produces, so this
package is a drop-in equivalent for code that was relying on either API.

The primary function is SHA256: it feeds the input through crypto/sha256 and
hex-encodes the resulting 32-byte digest with encoding/hex, yielding a
64-character lowercase string. The package also mirrors the companion
algorithms those JS libraries ship, offering SHA-1 (40 hex characters) and MD5
(32 hex characters), plus keyed HMAC-SHA256. Every function has a "String"
convenience variant that accepts a string argument instead of a []byte; the
String form simply converts its argument with []byte(s) and delegates, so the
two forms are exact equivalents for UTF-8 text.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/sha256hex"
```

## Usage

This is the package's own `ExampleHMACSHA256String`, so it compiles and its output is
asserted on every `go test ./sha256hex/`.

```go
fmt.Println(sha256hex.HMACSHA256String("key", "The quick brown fox jumps over the lazy dog"))
```

```
f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func HMACSHA256(key, data []byte) string` | HMACSHA256 returns the lowercase hex HMAC-SHA256 of data using key. |
| `func HMACSHA256String(key, data string) string` | HMACSHA256String returns the lowercase hex HMAC-SHA256 of data using key. |
| `func MD5(data []byte) string` | MD5 returns the lowercase hex MD5 digest of data. |
| `func MD5String(s string) string` | MD5String returns the lowercase hex MD5 digest of s. |
| `func SHA1(data []byte) string` | SHA1 returns the lowercase hex SHA-1 digest of data. |
| `func SHA1String(s string) string` | SHA1String returns the lowercase hex SHA-1 digest of s. |
| `func SHA256(data []byte) string` | SHA256 returns the lowercase hex SHA-256 digest of data. |
| `func SHA256String(s string) string` | SHA256String returns the lowercase hex SHA-256 digest of s. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/sha256hex).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
