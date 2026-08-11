# validatorjs — standalone port of the popular npm "validatorjs" (and the closely related

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/validatorjs.svg)](https://pkg.go.dev/github.com/malcolmston/express/validatorjs)

Package validatorjs is a standalone port of the popular npm "validatorjs" (and
the closely related "validator.js") string validation library to idiomatic Go.
It depends only on the Go standard library and does not import any part of the
express package, so it can be used from any program that needs to answer
simple "is this string a valid X?" questions.

Reach for this package whenever you have a string of unknown provenance - form
input, a query parameter, a config value, a webhook payload - and you need a
cheap, allocation-light predicate that tells you whether it is a valid e-mail,
URL, UUID, credit-card number, IP address, and so on. Every exported function
takes a single string and returns a bool: true when the input satisfies the
rule and false otherwise. There are no error returns and no panics on ordinary
input, which makes the functions convenient to compose inside larger
validation logic or to pass around as func(string) bool values.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/validatorjs"
```

## Usage

This is the package's own `ExampleIsCreditCard`, so it compiles and its output is
asserted on every `go test ./validatorjs/`.

```go
fmt.Println(validatorjs.IsCreditCard("4111 1111 1111 1111"))
	fmt.Println(validatorjs.IsCreditCard("4111 1111 1111 1112"))
```

```
true
false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Contains(s, substr string) bool` | Contains reports whether s contains the substring substr. |
| `func IsAlpha(s string) bool` | IsAlpha reports whether s is non-empty and contains only ASCII letters (a-z, A-Z). |
| `func IsAlphanumeric(s string) bool` | IsAlphanumeric reports whether s is non-empty and contains only ASCII letters and digits. |
| `func IsBase64(s string) bool` | IsBase64 reports whether s is a valid standard (RFC 4648) base64 string. |
| `func IsCreditCard(s string) bool` | IsCreditCard reports whether s is a valid credit card number. |
| `func IsEmail(s string) bool` | IsEmail reports whether s is a valid e-mail address. |
| `func IsFloat(s string) bool` | IsFloat reports whether s is a valid floating-point number, allowing an optional sign, a decimal point, and scientific notation. |
| `func IsHexColor(s string) bool` | IsHexColor reports whether s is a valid hexadecimal color, requiring a leading "#" and either three or six hexadecimal digits. |
| `func IsIP(s string) bool` | IsIP reports whether s is a valid IPv4 or IPv6 address. |
| `func IsIPv4(s string) bool` | IsIPv4 reports whether s is a valid IPv4 address in dotted-decimal notation. |
| `func IsIPv6(s string) bool` | IsIPv6 reports whether s is a valid IPv6 address. |
| `func IsInt(s string) bool` | IsInt reports whether s is a valid integer with an optional leading sign and no superfluous leading zeros. |
| `func IsJSON(s string) bool` | IsJSON reports whether s is a valid JSON document. |
| `func IsJWT(s string) bool` | IsJWT reports whether s is a well-formed JSON Web Token: three non-empty base64url-encoded segments separated by dots. |
| `func IsLength(s string, min, max int) bool` | IsLength reports whether the number of runes in s is between min and max inclusive. |
| `func IsMobilePhone(s string) bool` | IsMobilePhone reports whether s looks like a mobile phone number: an optional leading "+" followed by 7 to 15 digits. |
| `func IsMongoId(s string) bool` | IsMongoId reports whether s is a valid MongoDB ObjectId: exactly 24 hexadecimal characters. |
| `func IsNumeric(s string) bool` | IsNumeric reports whether s consists only of digits with an optional leading sign. |
| `func IsSemVer(s string) bool` | IsSemVer reports whether s is a valid Semantic Versioning 2.0.0 version string as defined at https://semver.org. |
| `func IsSlug(s string) bool` | IsSlug reports whether s is a valid slug: lowercase alphanumeric words separated by single hyphens, with no leading or trailing hyphen. |
| `func IsStrongPassword(s string) bool` | IsStrongPassword reports whether s is a strong password: at least eight characters including at least one lowercase letter, one uppercase letter, one… |
| `func IsURL(s string) bool` | IsURL reports whether s is a valid URL with an http, https, ftp, or ftps scheme and a host component. |
| `func IsUUID(s string) bool` | IsUUID reports whether s is a valid UUID of any version (1-5), matching the canonical 8-4-4-4-12 hexadecimal format. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/validatorjs).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
