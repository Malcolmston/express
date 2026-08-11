# nodepath — standard-library-only Go port of Node.js's POSIX "path" module

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/nodepath.svg)](https://pkg.go.dev/github.com/malcolmston/express/nodepath)

Package nodepath is a standard-library-only Go port of Node.js's POSIX "path"
module (https://nodejs.org/api/path.html), the utility Express and virtually
every Node program uses to manipulate filesystem path strings. It reproduces
the exact string semantics of node:path/posix — Basename, Dirname, Extname,
Join, Normalize, Resolve, Relative, IsAbsolute, Parse and Format — so path
logic can be ported from JavaScript to Go without behavioural surprises.

The functions are chosen because Go's own path and path/filepath packages,
while similar, differ from Node in edge cases that matter when porting code:
Node's Extname treats a leading dot as part of the name (Extname(".bashrc") is
""), Node's Parse reports an empty Dir for a bare filename (whereas Dirname
returns "."), Normalize preserves a trailing slash, and Join returns "." for
empty input. Matching those rules exactly is the point of this package.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/nodepath"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Basename(p string) string` | Basename returns the last portion of a path, with any trailing separators removed. |
| `func BasenameExt(p, ext string) string` | BasenameExt returns the last portion of a path with the given suffix removed when the basename ends with it. |
| `func Dirname(p string) string` | Dirname returns the directory portion of a path (everything before the last segment). |
| `func Extname(p string) string` | Extname returns the extension of the last path segment, from the last "." to the end. |
| `func Format(pp ParsedPath) string` | Format builds a path string from a ParsedPath, the inverse of Parse. |
| `func IsAbsolute(p string) bool` | IsAbsolute reports whether p is an absolute POSIX path (begins with "/"). |
| `func Join(elem ...string) string` | Join joins path segments with the POSIX separator and normalizes the result. |
| `func Normalize(p string) string` | Normalize normalizes a POSIX path, resolving ".." and "." segments and collapsing repeated separators. |
| `func Relative(from, to string) string` | Relative returns the relative path from from to to. |
| `func Resolve(elem ...string) string` | Resolve resolves a sequence of path segments into an absolute path, processing from right to left until an absolute path is built; if the segments… |

### Types

| Type | What it is |
| --- | --- |
| `ParsedPath` | ParsedPath is the structured decomposition of a path returned by Parse: Root is the filesystem root ("/" or ""), Dir is the full directory, Base is… |

<details>
<summary><code>ParsedPath</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Parse(p string) ParsedPath` | Parse decomposes a path into a ParsedPath. |

</details>

### Constants

`Delimiter`, `Sep`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/nodepath).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
