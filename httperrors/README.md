# httperrors — provides an HTTP-aware error type modeled on the npm "http-errors" package

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/httperrors.svg)](https://pkg.go.dev/github.com/malcolmston/express/httperrors)

Package httperrors provides an HTTP-aware error type modeled on the npm
"http-errors" package, the module used internally by Express and Koa to build
the errors that middleware turns into HTTP responses. Errors carry an HTTP
status code, a message, and an "expose" flag indicating whether the message is
safe to send to clients, so a single value can flow from the point of failure
all the way to the response writer while carrying everything the handler needs
to render it.

The Node http-errors package became the de facto standard because it lets
application code fail with intent: instead of returning a bare error and
separately deciding on a status code, a handler throws createError(404) or new
NotFound() and the framework's error middleware reads the status straight off
the error. This port brings that ergonomics to Go's error model. Error
implements the standard error interface, so a *Error can be returned anywhere
an error is expected, passed through errors.Is/As-style checks, and later
recognized with IsHTTPError to recover its status and expose flag when writing
the response.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/httperrors"
```

## Usage

This is the package's own `ExampleIsHTTPError`, so it compiles and its output is
asserted on every `go test ./httperrors/`.

```go
fmt.Println(httperrors.IsHTTPError(httperrors.NotFound("")))
	fmt.Println(httperrors.IsHTTPError(errors.New("plain")))
```

```
true
false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func IsHTTPError(err error) bool` | IsHTTPError reports whether err is (or wraps) an *Error from this package. |

### Types

| Type | What it is |
| --- | --- |
| `Error` | Error is an error carrying an HTTP status code. |

<details>
<summary><code>Error</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func BadGateway(msg string) *Error` | BadGateway returns a 502 Bad Gateway error. |
| `func BadRequest(msg string) *Error` | BadRequest returns a 400 Bad Request error. |
| `func Conflict(msg string) *Error` | Conflict returns a 409 Conflict error. |
| `func Forbidden(msg string) *Error` | Forbidden returns a 403 Forbidden error. |
| `func GatewayTimeout(msg string) *Error` | GatewayTimeout returns a 504 Gateway Timeout error. |
| `func Gone(msg string) *Error` | Gone returns a 410 Gone error. |
| `func InternalServerError(msg string) *Error` | InternalServerError returns a 500 Internal Server Error error. |
| `func LengthRequired(msg string) *Error` | LengthRequired returns a 411 Length Required error. |
| `func MethodNotAllowed(msg string) *Error` | MethodNotAllowed returns a 405 Method Not Allowed error. |
| `func New(status int, msg string) *Error` | New creates a new *Error with the given status code and message. |
| `func NotAcceptable(msg string) *Error` | NotAcceptable returns a 406 Not Acceptable error. |
| `func NotFound(msg string) *Error` | NotFound returns a 404 Not Found error. |
| `func NotImplemented(msg string) *Error` | NotImplemented returns a 501 Not Implemented error. |
| `func PayloadTooLarge(msg string) *Error` | PayloadTooLarge returns a 413 Payload Too Large error. |
| `func PreconditionFailed(msg string) *Error` | PreconditionFailed returns a 412 Precondition Failed error. |
| `func RequestTimeout(msg string) *Error` | RequestTimeout returns a 408 Request Timeout error. |
| `func ServiceUnavailable(msg string) *Error` | ServiceUnavailable returns a 503 Service Unavailable error. |
| `func TooManyRequests(msg string) *Error` | TooManyRequests returns a 429 Too Many Requests error. |
| `func Unauthorized(msg string) *Error` | Unauthorized returns a 401 Unauthorized error. |
| `func UnsupportedMediaType(msg string) *Error` | UnsupportedMediaType returns a 415 Unsupported Media Type error. |
| `func (e *Error) Error() string` | Error implements the error interface and returns the message. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/httperrors).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
