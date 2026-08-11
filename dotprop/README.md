# dotprop — provides get/set/has/delete access to nested map`string`any structures using dotted

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/dotprop.svg)](https://pkg.go.dev/github.com/malcolmston/express/dotprop)

Package dotprop provides get/set/has/delete access to nested map`string`any
structures using dotted path strings, mirroring the behavior of the npm
"dot-prop" library, using only the Go standard library. It exposes four
functions — Get, Has, Set, and Delete — that address a value deep inside a
tree of maps and slices by a single string path such as "a.b.c" rather than by
a chain of hand-written map lookups and type assertions.

This is the utility you want when working with dynamically shaped data:
configuration loaded from JSON or YAML, a decoded request body, or any
map`string`any whose structure is known by path rather than by a Go type. It
lets a caller read config.Get(cfg, "server.tls.enabled") or set a deeply
nested default without first checking that every intermediate map exists,
which keeps option-merging and template-context code short and free of nil-map
panics.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/dotprop"
```

## Usage

This is the package's own `ExampleGet`, so it compiles and its output is
asserted on every `go test ./dotprop/`.

```go
obj := map[string]any{
		"a": map[string]any{
			"b": map[string]any{"c": 42},
		},
	}
	value, ok := dotprop.Get(obj, "a.b.c")
	fmt.Println(value, ok)
```

```
42 true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Delete(obj map[string]any, path string) bool` | Delete removes the value at the given dotted path. |
| `func Get(obj map[string]any, path string) (any, bool)` | Get retrieves the value at the given dotted path. |
| `func Has(obj map[string]any, path string) bool` | Has reports whether the given dotted path resolves to a value in obj. |
| `func Set(obj map[string]any, path string, value any) map[string]any` | Set assigns value at the given dotted path, creating intermediate map`string`any nodes as needed. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/dotprop).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
