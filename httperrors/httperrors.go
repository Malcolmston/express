// Package httperrors provides an HTTP-aware error type modeled on the
// npm "http-errors" package. Errors carry an HTTP status code, a message,
// and an "expose" flag indicating whether the message is safe to send to
// clients.
package httperrors

import "net/http"

// Error is an error carrying an HTTP status code.
//
// Status is the HTTP status code. Message is the human readable error
// message returned by Error. Expose reports whether the message is safe to
// expose to clients; by convention it defaults to true for 4xx client
// errors and false for 5xx server errors.
type Error struct {
	Status  int
	Message string
	Expose  bool
}

// Error implements the error interface and returns the message.
func (e *Error) Error() string {
	return e.Message
}

// New creates a new *Error with the given status code and message.
//
// If msg is empty the standard status text for the code is used. Expose
// defaults to true for 4xx status codes and false for 5xx status codes.
func New(status int, msg string) *Error {
	if msg == "" {
		msg = http.StatusText(status)
	}
	return &Error{
		Status:  status,
		Message: msg,
		Expose:  status >= 400 && status < 500,
	}
}

// IsHTTPError reports whether err is (or wraps) an *Error from this package.
func IsHTTPError(err error) bool {
	_, ok := err.(*Error)
	return ok
}

// BadRequest returns a 400 Bad Request error.
func BadRequest(msg string) *Error { return New(http.StatusBadRequest, msg) }

// Unauthorized returns a 401 Unauthorized error.
func Unauthorized(msg string) *Error { return New(http.StatusUnauthorized, msg) }

// Forbidden returns a 403 Forbidden error.
func Forbidden(msg string) *Error { return New(http.StatusForbidden, msg) }

// NotFound returns a 404 Not Found error.
func NotFound(msg string) *Error { return New(http.StatusNotFound, msg) }

// MethodNotAllowed returns a 405 Method Not Allowed error.
func MethodNotAllowed(msg string) *Error { return New(http.StatusMethodNotAllowed, msg) }

// NotAcceptable returns a 406 Not Acceptable error.
func NotAcceptable(msg string) *Error { return New(http.StatusNotAcceptable, msg) }

// RequestTimeout returns a 408 Request Timeout error.
func RequestTimeout(msg string) *Error { return New(http.StatusRequestTimeout, msg) }

// Conflict returns a 409 Conflict error.
func Conflict(msg string) *Error { return New(http.StatusConflict, msg) }

// Gone returns a 410 Gone error.
func Gone(msg string) *Error { return New(http.StatusGone, msg) }

// LengthRequired returns a 411 Length Required error.
func LengthRequired(msg string) *Error { return New(http.StatusLengthRequired, msg) }

// PreconditionFailed returns a 412 Precondition Failed error.
func PreconditionFailed(msg string) *Error { return New(http.StatusPreconditionFailed, msg) }

// PayloadTooLarge returns a 413 Payload Too Large error.
func PayloadTooLarge(msg string) *Error { return New(http.StatusRequestEntityTooLarge, msg) }

// UnsupportedMediaType returns a 415 Unsupported Media Type error.
func UnsupportedMediaType(msg string) *Error { return New(http.StatusUnsupportedMediaType, msg) }

// TooManyRequests returns a 429 Too Many Requests error.
func TooManyRequests(msg string) *Error { return New(http.StatusTooManyRequests, msg) }

// InternalServerError returns a 500 Internal Server Error error.
func InternalServerError(msg string) *Error { return New(http.StatusInternalServerError, msg) }

// NotImplemented returns a 501 Not Implemented error.
func NotImplemented(msg string) *Error { return New(http.StatusNotImplemented, msg) }

// BadGateway returns a 502 Bad Gateway error.
func BadGateway(msg string) *Error { return New(http.StatusBadGateway, msg) }

// ServiceUnavailable returns a 503 Service Unavailable error.
func ServiceUnavailable(msg string) *Error { return New(http.StatusServiceUnavailable, msg) }

// GatewayTimeout returns a 504 Gateway Timeout error.
func GatewayTimeout(msg string) *Error { return New(http.StatusGatewayTimeout, msg) }
