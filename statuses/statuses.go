// Package statuses maps between HTTP status codes and their standard reason
// phrases and classifies codes by behavior. It is a Go port of the npm
// "statuses" package, reimplemented using only the Go standard library. Where
// the Node original exposes a callable module plus lookup tables, this package
// exposes the same capabilities as ordinary functions and returns Go-idiomatic
// values and errors.
//
// A web framework or HTTP client reaches for this package whenever it needs the
// canonical text for a numeric status code (for example, to build a default
// response body or a log line reading "404 Not Found"), or the reverse: to turn
// a human-written phrase such as "Not Found" back into the number 404. It also
// answers the three behavioral questions that middleware most often asks about a
// status: is it a redirect, may the request be retried, and must the response be
// sent without a body.
//
// Internally the package is backed by a code-to-message map covering the common
// registered codes in the 100-511 range and a lazily built, lower-cased
// message-to-code map for the reverse direction. Message performs a direct map
// lookup and returns the empty string for an unknown code. Code trims and
// lower-cases its input before looking it up, so matching is case-insensitive
// and tolerant of surrounding whitespace; an unknown phrase yields a non-nil
// error rather than a zero code that could be mistaken for a valid status.
//
// Three small sets drive classification. IsRedirect reports true for the 3xx
// codes that carry a Location header and cause the client to follow it (300,
// 301, 302, 303, 305, 307, 308); note that 304 Not Modified is deliberately
// excluded because it is a cache-validation response, not a redirect. IsRetry
// reports true for the gateway-family codes 502, 503, and 504, which typically
// represent a transient upstream failure that a client may safely retry. IsEmpty
// reports true for 204, 205, and 304, whose responses must never include a
// message body. These sets mirror the classification tables of the npm original.
//
// Codes returns every known status code sorted in ascending order, which is
// convenient for iterating over the full table (for instance to validate a
// round trip between Message and Code). Compared with the Node package, the data
// tables and classification rules are kept in parity, while the API surface is
// adapted to Go conventions: functions instead of a callable with attached
// properties, an (int, error) return from Code instead of a thrown exception,
// and a plain []int from Codes instead of an array of strings.
package statuses

import (
	"fmt"
	"sort"
	"strings"
)

// codeToMessage maps HTTP status codes to their standard reason phrases,
// covering common codes in the 100-511 range.
var codeToMessage = map[int]string{
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

// messageToCode is the reverse mapping, keyed by lower-cased reason phrase.
var messageToCode = func() map[string]int {
	m := make(map[string]int, len(codeToMessage))
	for code, msg := range codeToMessage {
		m[strings.ToLower(msg)] = code
	}
	return m
}()

// redirectCodes is the set of status codes considered redirects.
var redirectCodes = map[int]bool{
	300: true, 301: true, 302: true, 303: true,
	305: true, 307: true, 308: true,
}

// retryCodes is the set of status codes that indicate the request may be
// retried.
var retryCodes = map[int]bool{
	502: true, 503: true, 504: true,
}

// emptyCodes is the set of status codes that must not include a response body.
var emptyCodes = map[int]bool{
	204: true, 205: true, 304: true,
}

// Message returns the reason phrase for the given status code. It returns
// an empty string if the code is unknown.
func Message(code int) string {
	return codeToMessage[code]
}

// Code returns the status code for the given reason phrase. Matching is
// case-insensitive. It returns an error if the message is unknown.
func Code(message string) (int, error) {
	if code, ok := messageToCode[strings.ToLower(strings.TrimSpace(message))]; ok {
		return code, nil
	}
	return 0, fmt.Errorf("invalid status message: %q", message)
}

// IsRedirect reports whether the status code is a redirect.
func IsRedirect(code int) bool { return redirectCodes[code] }

// IsRetry reports whether a request that received this status code may be
// retried.
func IsRetry(code int) bool { return retryCodes[code] }

// IsEmpty reports whether responses with this status code must not carry a
// body.
func IsEmpty(code int) bool { return emptyCodes[code] }

// Codes returns all known status codes in ascending order.
func Codes() []int {
	codes := make([]int, 0, len(codeToMessage))
	for code := range codeToMessage {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	return codes
}
