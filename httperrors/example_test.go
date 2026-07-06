package httperrors_test

import (
	"errors"
	"fmt"

	"github.com/malcolmston/express/httperrors"
)

// ExampleNew builds an error from a status code and message. When the message
// is empty, New substitutes the standard reason phrase for the code, so
// New(404, "") produces the message "Not Found". The Expose flag is derived
// from the status class: 4xx codes default to exposable because client errors
// are generally safe to describe. The returned *Error satisfies the error
// interface, and its Error method returns the message. This makes New a
// one-line way to create a fully populated HTTP error.
func ExampleNew() {
	e := httperrors.New(404, "")
	fmt.Println(e.Status)
	fmt.Println(e.Error())
	fmt.Println(e.Expose)
	// Output:
	// 404
	// Not Found
	// true
}

// ExampleNotFound shows one of the named constructors, which are thin wrappers
// around New using the matching net/http status constant. Passing an empty
// message falls back to the canonical status text, here "Not Found". Like all
// 4xx helpers, NotFound marks the error as exposable. These constructors read
// clearly at the call site, letting handlers fail with intent such as
// returning httperrors.NotFound(""). The resulting value carries the 404 status
// for middleware to translate into a response.
func ExampleNotFound() {
	e := httperrors.NotFound("")
	fmt.Println(e.Status)
	fmt.Println(e.Message)
	fmt.Println(e.Expose)
	// Output:
	// 404
	// Not Found
	// true
}

// ExampleInternalServerError demonstrates the 5xx behavior around the Expose
// flag. Server errors may contain internal detail, so their message is not
// marked safe to send to clients: Expose defaults to false for any 5xx code.
// The default message still comes from the standard status text when none is
// supplied. This lets a handler signal a 500 without accidentally leaking
// implementation details to the caller. Applications that want to surface a
// curated message can set the Expose field on the returned error directly.
func ExampleInternalServerError() {
	e := httperrors.InternalServerError("")
	fmt.Println(e.Status)
	fmt.Println(e.Message)
	fmt.Println(e.Expose)
	// Output:
	// 500
	// Internal Server Error
	// false
}

// ExampleBadRequest illustrates supplying a custom message to a named
// constructor. When a non-empty message is given it is used verbatim instead of
// the default status text, so the error reads exactly as written. The status is
// still fixed by the constructor, here 400. Custom messages are handy for
// validation failures where the specific problem should be reported. Because
// this is a 4xx error, the message remains exposable to the client by default.
func ExampleBadRequest() {
	e := httperrors.BadRequest("email is required")
	fmt.Println(e.Status)
	fmt.Println(e.Error())
	// Output:
	// 400
	// email is required
}

// ExampleIsHTTPError shows how to recognize errors produced by this package.
// IsHTTPError reports whether a value returned as a plain error is actually an
// *Error, which lets response-writing code recover the status and expose flag.
// A value created by New or any named constructor reports true, while an
// ordinary error created elsewhere reports false. This is the typical bridge
// between application code that returns errors and middleware that must decide
// on a status code. Here a NotFound is recognized while a plain error is not.
func ExampleIsHTTPError() {
	fmt.Println(httperrors.IsHTTPError(httperrors.NotFound("")))
	fmt.Println(httperrors.IsHTTPError(errors.New("plain")))
	// Output:
	// true
	// false
}
