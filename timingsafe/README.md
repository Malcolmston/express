# timingsafe — compares byte slices and strings in constant time to avoid leaking secrets through

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/timingsafe.svg)](https://pkg.go.dev/github.com/malcolmston/express/timingsafe)

Package timingsafe compares byte slices and strings in constant time to avoid
leaking secrets through timing side channels. It is a stdlib-only Go port of
Node's crypto.timingSafeEqual
(https://nodejs.org/api/crypto.html#cryptotimingsafeequala-b), which is also
republished on npm as the "timing-safe-equal" package. The problem it solves
is that the obvious way to compare two byte strings — a byte-by-byte loop
that stops at the first difference — takes longer the more leading bytes
match, so an attacker who can measure comparison time can recover a secret one
byte at a time.

The canonical place this matters is verifying secrets supplied by an untrusted
caller against a value the server holds: HMAC signatures, session tokens,
password-reset tokens, API keys, CSRF tokens, and one-time codes. For any such
check the comparison must take the same amount of time whether the candidate
is completely wrong or wrong only in its final byte. Using an ordinary == or
bytes.Equal for these comparisons is a real, exploitable vulnerability; this
package exists so that the safe comparison is a single obvious call.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/timingsafe"
```

## Usage

This is the package's own `ExampleEqual`, so it compiles and its output is
asserted on every `go test ./timingsafe/`.

```go
fmt.Println(timingsafe.Equal([]byte("secret"), []byte("secret")))
	fmt.Println(timingsafe.Equal([]byte("secret"), []byte("sekret")))
	fmt.Println(timingsafe.Equal([]byte("secret"), []byte("secretx")))
```

```
true
false
false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Equal(a, b []byte) bool` | Equal reports whether a and b are equal using a constant-time comparison. |
| `func EqualStrict(a, b []byte) (bool, error)` | EqualStrict reports whether a and b are equal using a constant-time comparison, returning ErrLengthMismatch instead of false when the lengths differ. |
| `func EqualString(a, b string) bool` | EqualString reports whether a and b are equal using a constant-time comparison. |
| `func EqualStringStrict(a, b string) (bool, error)` | EqualStringStrict is EqualStrict for strings. |

### Variables

`ErrLengthMismatch`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/timingsafe).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
