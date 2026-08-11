# debounce — provides a faithful port of lodash's debounce utility

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/debounce.svg)](https://pkg.go.dev/github.com/malcolmston/express/debounce)

Package debounce provides a faithful port of lodash's debounce utility, built
on only the Go standard library. A debounced function coalesces a burst of
rapid calls into a single deferred invocation of the wrapped function, firing
it only after activity has quieted down. The package exposes a Debouncer
created by New, whose Call method drives the debounced behavior and whose
Cancel, Flush, and Pending methods manage the pending invocation.

Debouncing is the tool you want whenever an event fires far more often than
you need to react to it: coalescing a stream of keystrokes into one search
request, collapsing a flurry of file-change notifications into a single
rebuild, rate-limiting a "save draft" call, or making sure an expensive
recomputation runs once after a resize storm rather than on every intermediate
step. It keeps the timing bookkeeping out of the caller so the surrounding
code only has to say "something happened" by calling Call.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/debounce"
```

## Usage

This is the package's own `ExampleDebouncer`, so it compiles and its output is
asserted on every `go test ./debounce/`.

```go
count := 0
	d := debounce.New(time.Hour, func() { count++ })

	d.Call()
	d.Call()
	d.Call()
	d.Flush()

	fmt.Println(count)
```

```
1
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Clock` | Clock abstracts time so that debouncing can be tested with a fake clock. |
| `Debouncer` | Debouncer is a debounced wrapper around a function. |
| `Option` | Option configures a Debouncer. |
| `Timer` | Timer represents a single scheduled call created by Clock.AfterFunc. |

<details>
<summary><code>Debouncer</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func New(wait time.Duration, fn func(), opts ...Option) *Debouncer` | New creates a Debouncer that delays calling fn until wait has elapsed since the last Call. |
| `func (d *Debouncer) Call()` | Call invokes the debounced behavior. |
| `func (d *Debouncer) Cancel()` | Cancel discards any pending trailing invocation. |
| `func (d *Debouncer) Flush()` | Flush immediately invokes any pending trailing call. |
| `func (d *Debouncer) Pending() bool` | Pending reports whether an invocation is scheduled. |

</details>

<details>
<summary><code>Option</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func WithClock(c Clock) Option` | WithClock injects a custom Clock, primarily for testing. |
| `func WithLeading(leading bool) Option` | WithLeading enables or disables invoking on the leading edge of the timeout. |
| `func WithMaxWait(maxWait time.Duration) Option` | WithMaxWait sets the maximum time fn is allowed to be delayed before it is invoked, mirroring lodash's maxWait option. |
| `func WithTrailing(trailing bool) Option` | WithTrailing enables or disables invoking on the trailing edge of the timeout. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/debounce).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
