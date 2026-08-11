# ipaddr — provides parsing and classification of IPv4 and IPv6 addresses

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/ipaddr.svg)](https://pkg.go.dev/github.com/malcolmston/express/ipaddr)

Package ipaddr provides parsing and classification of IPv4 and IPv6 addresses.
It is a small Go port of a subset of the JavaScript "ipaddr.js" library, built
entirely on top of the standard library net package so that no third-party
dependency is required.

The ipaddr.js module is widely used in the Node ecosystem (notably by Express
and proxy middleware) to answer questions such as "is this a valid address?",
"which family is it?", "does it fall inside this CIDR block?" and "what kind
of address is it — loopback, private, multicast, public unicast?". This port
exists to answer the same questions in Go programs that need to reason about
client addresses, trust proxies, or filter traffic, while keeping the small,
object-oriented feel of the original API: Parse yields an *Address value whose
methods (Kind, String, Match, Range) mirror the JavaScript instance methods.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/ipaddr"
```

## Usage

This is the package's own `ExampleAddress_Match`, so it compiles and its output is
asserted on every `go test ./ipaddr/`.

```go
a, _ := ipaddr.Parse("192.168.5.5")
	in, _ := a.Match("192.168.0.0/16")
	out, _ := a.Match("10.0.0.0/8")
	fmt.Println(in)
	fmt.Println(out)
```

```
true
false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func IsValid(s string) bool` | IsValid reports whether s is a valid IPv4 or IPv6 address. |

### Types

| Type | What it is |
| --- | --- |
| `Address` | Address represents a parsed IPv4 or IPv6 address. |

<details>
<summary><code>Address</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Parse(s string) (*Address, error)` | Parse parses s as an IPv4 or IPv6 address and returns an *Address. |
| `func (a *Address) Kind() string` | Kind returns the address family: "ipv4" or "ipv6". |
| `func (a *Address) Match(cidr string) (bool, error)` | Match reports whether the address is contained within the given CIDR range, for example "192.168.0.0/16" or "2001:db8::/32". |
| `func (a *Address) Range() string` | Range classifies the address, returning a category name. |
| `func (a *Address) String() string` | String returns the canonical textual representation of the address. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/ipaddr).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
