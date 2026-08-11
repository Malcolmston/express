# keygrip — provides signing and verification of data using a rotating list of secret keys

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/keygrip.svg)](https://pkg.go.dev/github.com/malcolmston/express/keygrip)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/keygrip)

Package keygrip provides signing and verification of data using a rotating
list of secret keys, a stdlib-only Go port of the npm "keygrip" library used
by frameworks such as Express and Koa to protect signed cookies and other
tamper-evident tokens. Rather than encrypting data, keygrip attaches a keyed
message-authentication code (a digest) that a later request can present
alongside the data so the server can confirm the data was produced by a holder
of one of its secret keys and has not been altered in transit.

The motivating problem is secret rotation. A long-lived service wants to
replace its signing secret periodically without invalidating every cookie or
token that is still in the wild. Keygrip solves this by holding an ordered
list of keys: the newest key sits at the front and signs all new data, while
the older keys are retained only so that previously issued digests still
verify. When it is time to rotate, the operator prepends a fresh key and
eventually drops the oldest one from the tail, and outstanding tokens signed
under the retiring key keep working until they expire.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/keygrip"
```

## Usage

This is the package's own `ExampleKeygrip`, so it compiles and its output is
asserted on every `go test ./keygrip/`.

```go
kg := keygrip.New([]string{"current-secret", "previous-secret"})

	digest := kg.Sign("user=42")
	fmt.Println(kg.Verify("user=42", digest))
	fmt.Println(kg.Index("user=42", digest))
```

```
true
0
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Keygrip` | Keygrip holds an ordered list of secret keys. |
| `Options` | Options configures the digest scheme, mirroring the second and third arguments of the Node constructor keygrip(keys, algorithm, encoding). |

<details>
<summary><code>Keygrip</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func New(keys []string) *Keygrip` | New creates a Keygrip from the given keys, signing with HMAC-SHA256 and unpadded base64url. |
| `func NewWithOptions(keys []string, opts Options) (*Keygrip, error)` | NewWithOptions creates a Keygrip with an explicit digest scheme, the equivalent of the Node constructor keygrip(keys, algorithm, encoding). |
| `func (k *Keygrip) Index(data, digest string) int` | Index returns the index of the first key that produces digest for data, using a constant-time comparison, or -1 if no key matches. |
| `func (k *Keygrip) Sign(data string) string` | Sign returns the digest of data computed with the current (first) key. |
| `func (k *Keygrip) Verify(data, digest string) bool` | Verify reports whether digest is a valid signature of data for any key. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/keygrip).

## Measured parity

Compared case-for-case against **`keygrip@1.1.0`** by the harness in
[`parity/express/nested/keygrip`](https://github.com/malcolmston/go/tree/main/parity/express/nested/keygrip):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 65 |
| Matching | 63 |
| Mismatching | 0 |
| Declared deviations | 2 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/keygrip/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
