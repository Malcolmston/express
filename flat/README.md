# flat — flattens nested maps into single-depth maps with delimited keys and unflattens them

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/flat.svg)](https://pkg.go.dev/github.com/malcolmston/express/flat)

Package flat flattens nested maps into single-depth maps with delimited keys
and unflattens them back, mirroring the npm "flat" library.

Flattening turns a tree such as {"a": {"b": 1}} into the single-level map
{"a.b": 1}, and unflattening reverses the process. This is convenient when a
nested structure must be projected onto a flat key space: environment
variables, form fields, dot-notation config, database columns, or any store
that only accepts string keys and scalar values.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/flat"
```

## Usage

This is the package's own `ExampleFlatten_customDelimiter`, so it compiles and its output is
asserted on every `go test ./flat/`.

```go
in := map[string]any{"a": map[string]any{"b": 1}}
	fmt.Println(flat.Flatten(in, flat.FlattenOpts{Delimiter: "/"}))
```

```
map[a/b:1]
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Flatten(m map[string]any, opts ...FlattenOpts) map[string]any` | Flatten recursively flattens a nested map into a single-level map whose keys are joined by the configured delimiter (default "."). |
| `func Unflatten(m map[string]any, opts ...UnflattenOpts) map[string]any` | Unflatten expands a single-level map with delimited keys back into a nested map structure. |

### Types

| Type | What it is |
| --- | --- |
| `FlattenOpts` | FlattenOpts configures Flatten. |
| `UnflattenOpts` | UnflattenOpts configures Unflatten. |

### Constants

`DefaultDelimiter`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/flat).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
