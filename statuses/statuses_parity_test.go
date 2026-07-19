package statuses

// Upstream-parity tests for the Go port of jshttp/statuses.
//
// Vectors are taken verbatim from the ORIGINAL npm library's own sources and
// test suite:
//   - https://raw.githubusercontent.com/jshttp/statuses/master/codes.json
//   - https://raw.githubusercontent.com/jshttp/statuses/master/test/test.js
//   - https://raw.githubusercontent.com/jshttp/statuses/master/index.js
//
// The Node original exposes a callable `status(code)` plus `.message`,
// `.code`, `.codes`, `.redirect`, `.empty`, and `.retry`. This Go port maps
// those capabilities to Message/Code/Codes/IsRedirect/IsEmpty/IsRetry. Where
// the Node function throws, the Go port either returns an error (Code) or the
// empty string (Message); the vectors below assert the equivalent behavior.

import "testing"

// codeMessagePairs mirrors codes.json exactly (the code<->message map).
// Source: https://raw.githubusercontent.com/jshttp/statuses/master/codes.json
var codeMessagePairs = map[int]string{
	100: "Continue",
	101: "Switching Protocols",
	102: "Processing",
	103: "Early Hints",
	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non-Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",
	207: "Multi-Status",
	208: "Already Reported",
	226: "IM Used",
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	304: "Not Modified",
	305: "Use Proxy",
	307: "Temporary Redirect",
	308: "Permanent Redirect",
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Payload Too Large",
	414: "URI Too Long",
	415: "Unsupported Media Type",
	416: "Range Not Satisfiable",
	417: "Expectation Failed",
	418: "I'm a Teapot",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	425: "Too Early",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	506: "Variant Also Negotiates",
	507: "Insufficient Storage",
	508: "Loop Detected",
	509: "Bandwidth Limit Exceeded",
	510: "Not Extended",
	511: "Network Authentication Required",
}

// TestParityMessageForCode asserts Message(code) returns the exact reason
// phrase from codes.json for every registered code, and that the reverse
// Code(message) round-trips.
// Source: codes.json + test/test.js "when given a number".
func TestParityMessageForCode(t *testing.T) {
	for code, want := range codeMessagePairs {
		if got := Message(code); got != want {
			t.Errorf("Message(%d) = %q, want %q", code, got, want)
		}
		got, err := Code(want)
		if err != nil {
			t.Errorf("Code(%q) unexpected error: %v", want, err)
			continue
		}
		if got != code {
			t.Errorf("Code(%q) = %d, want %d", want, got, code)
		}
	}
}

// TestParityMessageSpotChecks mirrors the explicit numeric vectors in
// test/test.js "should return message when a valid status code".
// Source: test/test.js.
func TestParityMessageSpotChecks(t *testing.T) {
	cases := map[int]string{
		200: "OK",
		404: "Not Found",
		500: "Internal Server Error",
	}
	for code, want := range cases {
		if got := Message(code); got != want {
			t.Errorf("Message(%d) = %q, want %q", code, got, want)
		}
	}
}

// TestParityCodeCaseInsensitive mirrors test/test.js "should be case
// insensitive": status('Ok'), status('not found'), status('INTERNAL SERVER
// ERROR') are all valid lookups.
// Source: test/test.js.
func TestParityCodeCaseInsensitive(t *testing.T) {
	cases := map[string]int{
		"Ok":                    200,
		"not found":             404,
		"INTERNAL SERVER ERROR": 500,
		// exact-case forms from "should be truthy when a valid status message"
		"OK":                    200,
		"Not Found":             404,
		"Internal Server Error": 500,
	}
	for msg, want := range cases {
		got, err := Code(msg)
		if err != nil {
			t.Errorf("Code(%q) unexpected error: %v", msg, err)
			continue
		}
		if got != want {
			t.Errorf("Code(%q) = %d, want %d", msg, got, want)
		}
	}
}

// TestParityCodeUnknownMessage mirrors test/test.js "should throw for unknown
// status message". These include prototype-pollution probes ('constructor',
// '__proto__') that must NOT resolve to a code.
// Source: test/test.js.
func TestParityCodeUnknownMessage(t *testing.T) {
	for _, msg := range []string{"too many bugs", "constructor", "__proto__"} {
		if got, err := Code(msg); err == nil {
			t.Errorf("Code(%q) = %d, nil; want error", msg, got)
		}
	}
}

// TestParityMessageUnknownCode mirrors test/test.js "should throw for invalid /
// unknown / discontinued status code" (0, 1000, 299, 310, 306). The Node
// original throws; the Go port returns the empty string for an unregistered
// code, which is the idiomatic equivalent of "no message".
// Source: test/test.js "when given a number".
func TestParityMessageUnknownCode(t *testing.T) {
	for _, code := range []int{0, 1000, 299, 310, 306} {
		if got := Message(code); got != "" {
			t.Errorf("Message(%d) = %q, want %q", code, got, "")
		}
	}
}

// TestParityRedirect mirrors index.js status.redirect and test/test.js
// ".redirect should include 308". 304 is deliberately NOT a redirect.
// Source: index.js + test/test.js.
func TestParityRedirect(t *testing.T) {
	for _, c := range []int{300, 301, 302, 303, 305, 307, 308} {
		if !IsRedirect(c) {
			t.Errorf("IsRedirect(%d) = false, want true", c)
		}
	}
	for _, c := range []int{304, 306, 200, 404} {
		if IsRedirect(c) {
			t.Errorf("IsRedirect(%d) = true, want false", c)
		}
	}
}

// TestParityEmpty mirrors index.js status.empty and test/test.js ".empty should
// include 204".
// Source: index.js + test/test.js.
func TestParityEmpty(t *testing.T) {
	for _, c := range []int{204, 205, 304} {
		if !IsEmpty(c) {
			t.Errorf("IsEmpty(%d) = false, want true", c)
		}
	}
	for _, c := range []int{200, 404} {
		if IsEmpty(c) {
			t.Errorf("IsEmpty(%d) = true, want false", c)
		}
	}
}

// TestParityRetry mirrors index.js status.retry and test/test.js ".retry should
// include 504".
// Source: index.js + test/test.js.
func TestParityRetry(t *testing.T) {
	for _, c := range []int{502, 503, 504} {
		if !IsRetry(c) {
			t.Errorf("IsRetry(%d) = false, want true", c)
		}
	}
	for _, c := range []int{429, 500, 200} {
		if IsRetry(c) {
			t.Errorf("IsRetry(%d) = true, want false", c)
		}
	}
}

// TestParityCodesContainsAll mirrors test/test.js ".codes should include codes
// from Node.js": every code in codes.json must appear in Codes().
// Source: test/test.js + codes.json.
func TestParityCodesContainsAll(t *testing.T) {
	set := make(map[int]bool, len(Codes()))
	for _, c := range Codes() {
		set[c] = true
	}
	for code := range codeMessagePairs {
		if !set[code] {
			t.Errorf("Codes() missing %d", code)
		}
	}
	if len(Codes()) != len(codeMessagePairs) {
		t.Errorf("len(Codes()) = %d, want %d", len(Codes()), len(codeMessagePairs))
	}
}
