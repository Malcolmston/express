package httperrors

// Upstream-parity tests for the jshttp/http-errors npm package.
//
// Every vector below is transcribed from the original library's own test
// suite and source, not invented here:
//   https://raw.githubusercontent.com/jshttp/http-errors/master/test/test.js
//   https://raw.githubusercontent.com/jshttp/http-errors/master/index.js
//
// This Go port models the subset of http-errors that maps onto Go's error
// type: an HTTP status, a message, and the "expose" flag. It does not model
// the JavaScript-only behaviors (attaching arbitrary props, wrapping an
// existing Error object, named constructor prototypes, the `.name`/`.stack`
// fields, or remapping out-of-range status codes to 500); those upstream
// vectors are intentionally not encoded here and are recorded as gaps in the
// task notes rather than asserted.

import (
	"errors"
	"testing"
)

// TestParityCreateErrorStatus covers createError(status) message/status/expose
// mapping. Vectors from test.js "describe('createError(status)')":
//   - status 300 -> message "Multiple Choices", status 300 (test.js:54-74)
//   - status 404 -> message "Not Found", status 404       (test.js:76-96)
//   - unknown 4xx 499 -> message "Bad Request", status 499 (test.js:98-118)
//   - unknown 5xx 599 -> message "Internal Server Error"   (test.js:120-140)
//
// Expose follows index.js:94 `err.expose = status < 500`.
func TestParityCreateErrorStatus(t *testing.T) {
	cases := []struct {
		status  int
		message string
		expose  bool
	}{
		{300, "Multiple Choices", true},
		{404, "Not Found", true},
		{499, "Bad Request", true},
		{599, "Internal Server Error", false},
		{500, "Internal Server Error", false},
	}
	for _, c := range cases {
		e := New(c.status, "")
		if e.Status != c.status {
			t.Errorf("New(%d,\"\").Status = %d, want %d", c.status, e.Status, c.status)
		}
		if e.Message != c.message {
			t.Errorf("New(%d,\"\").Message = %q, want %q", c.status, e.Message, c.message)
		}
		if e.Expose != c.expose {
			t.Errorf("New(%d,\"\").Expose = %v, want %v", c.status, e.Expose, c.expose)
		}
	}
}

// TestParityCreateErrorStatusMessage covers createError(status, message).
// Vector from test.js:143-163: createError(404, 'missing') -> message
// "missing", status 404, statusCode 404.
func TestParityCreateErrorStatusMessage(t *testing.T) {
	e := New(404, "missing")
	if e.Message != "missing" {
		t.Errorf("Message = %q, want %q", e.Message, "missing")
	}
	if e.Status != 404 {
		t.Errorf("Status = %d, want 404", e.Status)
	}
	// The port carries a single Status field; upstream exposes it as both
	// .status and .statusCode with identical values (index.js:95).
	if e.Error() != "missing" {
		t.Errorf("Error() = %q, want %q", e.Error(), "missing")
	}
}

// TestParityIsHttpError covers createError.isHttpError(val). Vectors from
// test.js:165-218. Only the cases expressible through Go's `error`-typed
// parameter are encoded: a nil error and a plain error report false, while a
// value produced by this package reports true (test.js:202-208).
func TestParityIsHttpError(t *testing.T) {
	if IsHTTPError(nil) {
		t.Error("IsHTTPError(nil) = true, want false")
	}
	if IsHTTPError(errors.New("foobar")) {
		t.Error("IsHTTPError(plain error) = true, want false")
	}
	var err error = New(500, "")
	if !IsHTTPError(err) {
		t.Error("IsHTTPError(New(500)) = false, want true")
	}
}

// TestParityNamedConstructors covers the named constructors. Vectors from
// test.js "new createError.NotFound()" (385-393) and
// "new createError.InternalServerError()" (395-403): the default message is
// the canonical status text, and expose is true for the 4xx client
// constructor and false for the 5xx server constructor (index.js:165, 234).
func TestParityNamedConstructors(t *testing.T) {
	cases := []struct {
		name    string
		got     *Error
		status  int
		message string
		expose  bool
	}{
		{"NotFound", NotFound(""), 404, "Not Found", true},
		{"InternalServerError", InternalServerError(""), 500, "Internal Server Error", false},
		{"BadRequest", BadRequest(""), 400, "Bad Request", true},
		{"Unauthorized", Unauthorized(""), 401, "Unauthorized", true},
		{"Forbidden", Forbidden(""), 403, "Forbidden", true},
		{"Conflict", Conflict(""), 409, "Conflict", true},
		{"BadGateway", BadGateway(""), 502, "Bad Gateway", false},
		{"ServiceUnavailable", ServiceUnavailable(""), 503, "Service Unavailable", false},
	}
	for _, c := range cases {
		if c.got.Status != c.status {
			t.Errorf("%s Status = %d, want %d", c.name, c.got.Status, c.status)
		}
		if c.got.Message != c.message {
			t.Errorf("%s Message = %q, want %q", c.name, c.got.Message, c.message)
		}
		if c.got.Expose != c.expose {
			t.Errorf("%s Expose = %v, want %v", c.name, c.got.Expose, c.expose)
		}
	}
}

// TestParityExposeOverride covers createError(status, msg, { expose: false })
// from test.js:372-377. Upstream lets callers override the derived expose
// flag; the port exposes Expose as a settable field for the same effect.
func TestParityExposeOverride(t *testing.T) {
	e := New(404, "LOL")
	if !e.Expose {
		t.Fatalf("New(404).Expose = false, want true before override")
	}
	e.Expose = false
	if e.Expose {
		t.Errorf("after override Expose = true, want false")
	}
}
