# proxyaddr — determines the real client address of a request that has passed through one or more

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/proxyaddr.svg)](https://pkg.go.dev/github.com/malcolmston/express/proxyaddr)

Package proxyaddr determines the real client address of a request that has
passed through one or more trusted reverse proxies. It is a Go port of the npm
module "proxy-addr", the same logic Express uses to implement its "trust
proxy" setting and the req.ip / req.ips accessors.

When an application sits behind a load balancer, CDN, or reverse proxy, the
socket's remote address is the address of the nearest proxy rather than the
end user. Each hop is expected to append the address it saw to the
X-Forwarded-For header, producing a left-to-right list that runs from the
original client through every intermediary. Because any of those values can be
forged by a client, they can only be trusted up to the point where the chain
leaves infrastructure you control. This package encodes that decision: given
the socket address, the header, and a predicate describing which hops are
trusted, it returns the closest address that is not itself a trusted proxy.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/proxyaddr"
```

## Usage

This is the package's own `ExampleProxyAddr`, so it compiles and its output is
asserted on every `go test ./proxyaddr/`.

```go
trust, err := proxyaddr.Compile([]string{"loopback"})
	if err != nil {
		panic(err)
	}
	fmt.Println(proxyaddr.ProxyAddr("127.0.0.1", "203.0.113.1", trust))
```

```
203.0.113.1
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func All(remoteAddr, xff string) []string` | All returns the full forwarded address chain for the request, starting with the socket remote address followed by the X-Forwarded-For entries in… |
| `func AllTrusted(remoteAddr, xff string, trust func(addr string, i int) bool) []string` | AllTrusted returns the forwarded address chain truncated at the first untrusted hop, mirroring proxy-addr's alladdrs(req, trust). |
| `func Compile(vals []string) (func(addr string, i int) bool, error)` | Compile builds a trust function from a list of trusted CIDR strings and/or the named presets "loopback", "linklocal", and "uniquelocal". |
| `func ProxyAddr(remoteAddr, xff string, trust func(addr string, i int) bool) string` | ProxyAddr walks the forwarded chain and returns the first address that is not trusted according to trust. |

### Types

| Type | What it is |
| --- | --- |
| `Trust` | Trust is the predicate ProxyAddr, All and AllTrusted consult to decide whether a hop in the forwarded chain is a proxy the application controls. |

<details>
<summary><code>Trust</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func CompileTrust(val any) (Trust, error)` | CompileTrust builds a Trust from the same loosely-typed value Express accepts for its "trust proxy" setting, mirroring Express's compileTrust: |
| `func TrustAll() Trust` | TrustAll returns a Trust that trusts every hop, so ProxyAddr resolves to the left-most X-Forwarded-For entry. |
| `func TrustHops(n int) Trust` | TrustHops returns a Trust that trusts the n hops nearest the application, regardless of their addresses. |
| `func TrustNone() Trust` | TrustNone returns a Trust that trusts nothing, so ProxyAddr always resolves to the socket remote address and AllTrusted returns just that address. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/proxyaddr).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
