// Package contentrange formats and parses HTTP Content-Range header values
// using only the Go standard library. It provides Format to build a header
// string from its numeric components, Parse to decode a header string into a
// ContentRange struct, and a String method to render a ContentRange back into a
// header value.
//
// A Content-Range header has the form "<unit> <range>/<size>", for example
// "bytes 0-499/1234". The range may be "*" for an unsatisfied range and the
// size may be "*" when the total length is unknown. Servers send this header on
// 206 Partial Content responses to describe which slice of a representation the
// body carries, and on 416 Range Not Satisfiable responses to report the total
// size; you use this package on the server to construct that header and on a
// client to interpret one.
//
// Format takes a unit and the start, end, and size as int64 values. An empty
// unit defaults to "bytes". A negative start signals an unsatisfied range and
// is rendered as "*" (the end is then ignored), and a negative size signals an
// unknown total and is rendered as "*". Otherwise the range is rendered as
// "start-end" and the size as its decimal value, so Format("bytes", 0, 499,
// 1234) returns "bytes 0-499/1234".
//
// Parse reverses that mapping. It trims surrounding whitespace, splits the unit
// off at the first space, then splits the remainder at the '/' into a range and
// a size. A "*" range clears HasRange and leaves Start and End at -1; a "*"
// size clears HasSize and leaves Size at -1. Concrete values are parsed as
// signed 64-bit integers and set the corresponding HasRange or HasSize flag to
// true. A missing unit, a missing '/' separator, a range without a '-', or any
// non-numeric number field yields a descriptive error and the zero
// ContentRange. Because the numeric fields double as sentinels, callers should
// consult HasRange and HasSize rather than testing Start, End, or Size against
// -1 directly.
//
// The design tracks the npm range-parser and Express partial-response
// conventions rather than mirroring one specific Node package byte for byte:
// the "*" handling for unsatisfied ranges and unknown sizes, and the
// unit/range/size grammar, match how those tools format the header. Round
// tripping is exact for the canonical forms, so Parse followed by String
// reproduces "bytes 0-499/1234", "bytes */1234", "bytes 0-499/*", and "bytes
// */*" unchanged.
package contentrange

import (
	"errors"
	"strconv"
	"strings"
)

// ContentRange represents the parsed components of a Content-Range header.
type ContentRange struct {
	// Unit is the range unit, typically "bytes".
	Unit string
	// Start and End are the inclusive byte range. They are meaningful only
	// when HasRange is true.
	Start int64
	End   int64
	// Size is the total length of the representation. It is meaningful only
	// when HasSize is true.
	Size int64
	// HasRange reports whether a concrete range (not "*") was present.
	HasRange bool
	// HasSize reports whether a concrete size (not "*") was present.
	HasSize bool
}

// Format builds a Content-Range header value.
//
// If unit is empty, "bytes" is used. A negative start produces an unsatisfied
// range ("*"), and a negative size produces an unknown size ("*"). For
// example, Format("bytes", 0, 499, 1234) returns "bytes 0-499/1234".
func Format(unit string, start, end, size int64) string {
	if unit == "" {
		unit = "bytes"
	}

	rangePart := "*"
	if start >= 0 {
		rangePart = strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end, 10)
	}

	sizePart := "*"
	if size >= 0 {
		sizePart = strconv.FormatInt(size, 10)
	}

	return unit + " " + rangePart + "/" + sizePart
}

// String renders the ContentRange back into a header value.
func (cr ContentRange) String() string {
	start, end, size := int64(-1), int64(-1), int64(-1)
	if cr.HasRange {
		start, end = cr.Start, cr.End
	}
	if cr.HasSize {
		size = cr.Size
	}
	return Format(cr.Unit, start, end, size)
}

// Parse parses a Content-Range header value such as "bytes 0-499/1234",
// "bytes */1234", or "bytes 0-499/*".
func Parse(s string) (ContentRange, error) {
	var cr ContentRange
	s = strings.TrimSpace(s)

	sp := strings.IndexByte(s, ' ')
	if sp <= 0 {
		return cr, errors.New("contentrange: missing unit")
	}
	cr.Unit = s[:sp]
	rest := s[sp+1:]

	slash := strings.IndexByte(rest, '/')
	if slash < 0 {
		return cr, errors.New("contentrange: missing size separator")
	}
	rangeStr := rest[:slash]
	sizeStr := rest[slash+1:]

	if rangeStr == "*" {
		cr.HasRange = false
		cr.Start, cr.End = -1, -1
	} else {
		dash := strings.IndexByte(rangeStr, '-')
		if dash < 0 {
			return cr, errors.New("contentrange: invalid range")
		}
		start, err := strconv.ParseInt(rangeStr[:dash], 10, 64)
		if err != nil {
			return cr, errors.New("contentrange: invalid range start")
		}
		end, err := strconv.ParseInt(rangeStr[dash+1:], 10, 64)
		if err != nil {
			return cr, errors.New("contentrange: invalid range end")
		}
		cr.Start, cr.End, cr.HasRange = start, end, true
	}

	if sizeStr == "*" {
		cr.HasSize = false
		cr.Size = -1
	} else {
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return cr, errors.New("contentrange: invalid size")
		}
		cr.Size, cr.HasSize = size, true
	}

	return cr, nil
}
