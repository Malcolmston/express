# deepmerge — deeply merges maps, modeled on the npm "deepmerge" package

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/deepmerge.svg)](https://pkg.go.dev/github.com/malcolmston/express/deepmerge)

Package deepmerge deeply merges maps, modeled on the npm "deepmerge" package.
It recursively combines a target map with a source map, producing a brand new
map in which nested objects are merged key by key rather than being replaced
wholesale. The public API is Merge for the common case, MergeWith for callers
that need to customize how slices combine, and MergeAll for folding an
arbitrary number of maps together left to right.

Reach for deepmerge whenever you need to layer configuration or state: default
settings overlaid by user settings, a base document patched by an override, or
several partial fragments assembled into one. Because later values win only at
the leaves where they actually appear, a source map that sets a single nested
field does not clobber sibling fields that only exist in the target. This is
the behavior that a shallow copy or a plain map assignment cannot give you.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/deepmerge"
```

## Usage

This is the package's own `ExampleMerge`, so it compiles and its output is
asserted on every `go test ./deepmerge/`.

```go
target := map[string]any{"a": 1, "nested": map[string]any{"x": 1}}
	source := map[string]any{"b": 2, "nested": map[string]any{"y": 2}}

	merged := deepmerge.Merge(target, source)
	fmt.Println(merged)
```

```
map[a:1 b:2 nested:map[x:1 y:2]]
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Merge(target, source map[string]any) map[string]any` | Merge returns a new map that is the deep merge of target and source. |
| `func MergeAll(maps ...map[string]any) map[string]any` | MergeAll deeply merges any number of maps left to right and returns a new map. |
| `func MergeWith(target, source map[string]any, opts Options) map[string]any` | MergeWith returns a new map that is the deep merge of target and source using opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures MergeWith. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/deepmerge).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
