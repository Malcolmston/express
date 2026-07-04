// Package contentrange formats and parses HTTP Content-Range header values.
//
// A Content-Range header has the form "<unit> <range>/<size>", for example
// "bytes 0-499/1234". The range may be "*" for an unsatisfied range and the
// size may be "*" when the total length is unknown.
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
