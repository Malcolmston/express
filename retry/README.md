# retry — retries a function with exponential backoff

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/retry.svg)](https://pkg.go.dev/github.com/malcolmston/express/retry)

Package retry retries a function with exponential backoff. It is a stdlib-only
Go port of the npm "retry" package (https://www.npmjs.com/package/retry, also
published as node-retry) and the higher-level "p-retry" / "async-retry"
wrappers built on top of it. Those JavaScript libraries repeatedly invoke an
operation that may fail transiently — a flaky network request, a database
call, a rate-limited API — pausing between attempts for a delay that grows
on each failure, and they are the retry engine behind a large part of the Node
ecosystem. This package brings the same retry-with-backoff contract to Go
using only the standard library.

The central entry point is Do, which calls fn and, on error, sleeps and calls
it again until the operation succeeds, the retry budget is exhausted, or an
AbortError signals that further attempts are pointless. fn receives the
1-based attempt number so it can log or vary behavior across attempts. When
every attempt fails, the error from the final attempt is returned unchanged;
when an AbortError is returned, its wrapped error (if any) is returned and no
further attempts are made. This mirrors p-retry, where returning an AbortError
short-circuits the retry loop.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/retry"
```

## Usage

This is the package's own `ExampleAbort`, so it compiles and its output is
asserted on every `go test ./retry/`.

```go
count := 0
	err := retry.Do(func(attempt int) error {
		count++
		return retry.Abort(errors.New("permanent failure"))
	}, retry.Options{Retries: 5, Sleep: func(time.Duration) {}})
	fmt.Println(count, err)
```

```
1 permanent failure
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Do(fn func(attempt int) error, opts Options) error` | Do calls fn, retrying with exponential backoff on error until it succeeds, the retry budget is exhausted, or an AbortError is returned. |

### Types

| Type | What it is |
| --- | --- |
| `AbortError` | AbortError wraps an error to signal that retrying must stop immediately. |
| `Options` | Options configures Do. |

<details>
<summary><code>AbortError</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Abort(err error) *AbortError` | Abort creates an AbortError that stops retrying immediately when returned from the retried function. |
| `func (a *AbortError) Error() string` | Error implements the error interface. |
| `func (a *AbortError) Unwrap() error` | Unwrap returns the wrapped error so errors.Is / errors.As work through it. |

</details>

<details>
<summary><code>Options</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func (o Options) Backoff(attempt int) time.Duration` | Backoff computes the delay before the given retry attempt. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/retry).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
