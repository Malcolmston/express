# humanizeduration — converts a duration in milliseconds into a human readable phrase such as "1 hour" or "1

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/humanizeduration.svg)](https://pkg.go.dev/github.com/malcolmston/express/humanizeduration)

Package humanizeduration converts a duration in milliseconds into a human
readable phrase such as "1 hour" or "1 minute, 1 second". It is a faithful
port of the npm package "humanize-duration" (English only), reproducing that
library's default behaviour and the handful of options most commonly reached
for when formatting durations for end users.

The original JavaScript module exists because raw millisecond counts are
unfriendly to read: "3661000" means little at a glance, whereas "1 hour, 1
minute, 1 second" is immediately understood. This port is intended for the
same job in Go programs, for example rendering an elapsed time, a countdown,
or a cache TTL in a log line, a CLI, or an HTTP response without pulling in a
third-party dependency. Only the standard library is used.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/humanizeduration"
```

## Usage

This is the package's own `ExampleHumanizeOpts`, so it compiles and its output is
asserted on every `go test ./humanizeduration/`.

```go
out := humanizeduration.HumanizeOpts(3661000, humanizeduration.Options{
		Largest:   2,
		Delimiter: " and ",
	})
	fmt.Println(out)
```

```
1 hour and 1 minute
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Humanize(ms int64) string` | Humanize converts a number of milliseconds into a human readable phrase using the default options. |
| `func HumanizeOpts(ms int64, opts Options) string` | HumanizeOpts converts a number of milliseconds into a human readable phrase honouring the supplied Options. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures how a duration is rendered. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/humanizeduration).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
