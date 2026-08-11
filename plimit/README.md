# plimit — provides a faithful port of the npm p-limit library

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/plimit.svg)](https://pkg.go.dev/github.com/malcolmston/express/plimit)

Package plimit provides a faithful port of the npm p-limit library: it runs an
unbounded number of scheduled functions while capping how many execute
concurrently at any instant. Where p-limit wraps promise-returning functions
and resolves them under a concurrency ceiling, this port wraps ordinary Go
funcs and runs them on goroutines under the same ceiling, using only the
standard library (sync and sync/atomic).

The motivation is the same one that makes p-limit popular in Node: you often
have many independent units of work — HTTP requests, file reads, database
queries — and firing all of them at once would exhaust sockets, file
handles, memory or a remote service's rate limit. A limiter lets you enqueue
every task up front for readability while guaranteeing that no more than N run
at the same moment. The remaining tasks queue and start automatically as
running ones finish.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/plimit"
```

## Usage

This is the package's own `ExampleLimiter`, so it compiles and its output is
asserted on every `go test ./plimit/`.

```go
l := plimit.New(2)
	var sum int64
	for i := 1; i <= 100; i++ {
		i := i
		l.Go(func() {
			atomic.AddInt64(&sum, int64(i))
		})
	}
	l.Wait()
	fmt.Println(sum)
```

```
5050
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Limiter` | Limiter runs scheduled functions with a bounded level of concurrency. |

<details>
<summary><code>Limiter</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func New(concurrency int) *Limiter` | New creates a Limiter that allows at most concurrency functions to run at the same time. |
| `func (l *Limiter) ActiveCount() int` | ActiveCount returns the number of functions currently executing. |
| `func (l *Limiter) Concurrency() int` | Concurrency returns the ceiling this Limiter was created with, after New's clamping of values below 1. |
| `func (l *Limiter) Go(fn func())` | Go schedules fn to run. |
| `func (l *Limiter) PendingCount() int` | PendingCount returns the number of functions scheduled but not yet started. |
| `func (l *Limiter) Run(fn func())` | Run schedules fn the same way as Go but blocks the caller until a concurrency slot is available before returning. |
| `func (l *Limiter) Wait()` | Wait blocks until all scheduled functions have finished executing. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/plimit).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
