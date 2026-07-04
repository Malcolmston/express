// Package statuses maps between HTTP status codes and their reason phrases,
// modeled on the npm "statuses" package. It also provides helpers for
// classifying status codes (redirects, retriable, empty).
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
