package httperrors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewDefaultsMessage(t *testing.T) {
	e := New(404, "")
	if e.Message != http.StatusText(404) {
		t.Fatalf("got message %q, want %q", e.Message, http.StatusText(404))
	}
	if e.Status != 404 {
		t.Fatalf("got status %d, want 404", e.Status)
	}
}

func TestNewCustomMessage(t *testing.T) {
	e := New(400, "bad input")
	if e.Error() != "bad input" {
		t.Fatalf("got %q, want %q", e.Error(), "bad input")
	}
}

func TestExposeFlag(t *testing.T) {
	if e := New(400, "x"); !e.Expose {
		t.Fatal("4xx should expose")
	}
	if e := New(500, "x"); e.Expose {
		t.Fatal("5xx should not expose")
	}
}

func TestIsHTTPError(t *testing.T) {
	var err error = New(404, "")
	if !IsHTTPError(err) {
		t.Fatal("expected IsHTTPError true")
	}
	if IsHTTPError(errors.New("plain")) {
		t.Fatal("expected IsHTTPError false for plain error")
	}
}

func TestHelpers(t *testing.T) {
	cases := []struct {
		got    *Error
		status int
	}{
		{BadRequest(""), 400},
		{Unauthorized(""), 401},
		{Forbidden(""), 403},
		{NotFound(""), 404},
		{MethodNotAllowed(""), 405},
		{NotAcceptable(""), 406},
		{RequestTimeout(""), 408},
		{Conflict(""), 409},
		{Gone(""), 410},
		{LengthRequired(""), 411},
		{PreconditionFailed(""), 412},
		{PayloadTooLarge(""), 413},
		{UnsupportedMediaType(""), 415},
		{TooManyRequests(""), 429},
		{InternalServerError(""), 500},
		{NotImplemented(""), 501},
		{BadGateway(""), 502},
		{ServiceUnavailable(""), 503},
		{GatewayTimeout(""), 504},
	}
	for _, c := range cases {
		if c.got.Status != c.status {
			t.Errorf("got status %d, want %d", c.got.Status, c.status)
		}
		if c.got.Message == "" {
			t.Errorf("status %d: empty default message", c.status)
		}
	}
}

func TestHelperCustomMessage(t *testing.T) {
	e := NotFound("nope")
	if e.Message != "nope" {
		t.Fatalf("got %q, want nope", e.Message)
	}
}
