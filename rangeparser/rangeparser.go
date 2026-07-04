// Package rangeparser parses the HTTP Range header, a port of the npm
// "range-parser" package. It resolves a Range header such as "bytes=0-499"
// against a total resource size and reports the byte ranges requested.
package rangeparser

import (
	"sort"
	"strconv"
	"strings"
)

// Range represents a single inclusive byte range [Start, End].
type Range struct {
	Start int64
	End   int64
}

// Result codes returned alongside the parsed ranges. A negative value
// indicates an error.
const (
	// ResultOK indicates the header was parsed successfully.
	ResultOK = 0
	// ResultUnsatisfiable indicates none of the ranges overlap the resource.
	ResultUnsatisfiable = -1
	// ResultMalformed indicates the header was malformed or used a bad unit.
	ResultMalformed = -2
)

// Ranges holds the parsed ranges together with the unit ("type") that was
// requested, e.g. "bytes".
type Ranges struct {
	Type   string
	Ranges []Range
}

// Parse parses the given Range header against a resource of the given size.
//
// It returns the slice of resolved inclusive ranges and a result code: 0 on
// success (ResultOK), -1 (ResultUnsatisfiable) when the ranges cannot be
// satisfied, and -2 (ResultMalformed) when the header is malformed or uses an
// unsupported unit. When combine is true, overlapping and adjacent ranges are
// merged.
func Parse(size int64, header string, combine bool) (ranges []Range, result int) {
	r, res := ParseRanges(size, header, combine)
	return r.Ranges, res
}

// ParseRanges is like Parse but also returns the requested unit via Ranges.Type.
func ParseRanges(size int64, header string, combine bool) (Ranges, int) {
	idx := strings.IndexByte(header, '=')
	if idx < 0 {
		return Ranges{}, ResultMalformed
	}

	unit := header[:idx]
	if unit == "" {
		return Ranges{}, ResultMalformed
	}

	out := Ranges{Type: unit}
	specs := strings.Split(header[idx+1:], ",")

	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		dash := strings.IndexByte(spec, '-')
		if dash < 0 {
			return Ranges{}, ResultMalformed
		}
		startStr := strings.TrimSpace(spec[:dash])
		endStr := strings.TrimSpace(spec[dash+1:])

		var start, end int64
		var err error

		if startStr == "" {
			// suffix range: -N means the last N bytes
			if endStr == "" {
				return Ranges{}, ResultMalformed
			}
			var suffix int64
			suffix, err = strconv.ParseInt(endStr, 10, 64)
			if err != nil {
				return Ranges{}, ResultMalformed
			}
			start = size - suffix
			end = size - 1
		} else {
			start, err = strconv.ParseInt(startStr, 10, 64)
			if err != nil {
				return Ranges{}, ResultMalformed
			}
			if endStr == "" {
				end = size - 1
			} else {
				end, err = strconv.ParseInt(endStr, 10, 64)
				if err != nil {
					return Ranges{}, ResultMalformed
				}
			}
		}

		// Clamp to the resource bounds.
		if end > size-1 {
			end = size - 1
		}
		if start < 0 {
			start = 0
		}

		// Skip nonsensical or entirely out-of-range specs.
		if start > end || start < 0 {
			continue
		}

		out.Ranges = append(out.Ranges, Range{Start: start, End: end})
	}

	if len(out.Ranges) == 0 {
		// No satisfiable range.
		return Ranges{Type: unit}, ResultUnsatisfiable
	}

	if combine {
		out.Ranges = combineRanges(out.Ranges)
	}

	return out, ResultOK
}

// combineRanges merges overlapping and adjacent ranges while preserving the
// original relative ordering of the first occurrence of each merged group.
func combineRanges(ranges []Range) []Range {
	type ordered struct {
		Range
		index int
	}
	ordinals := make([]ordered, len(ranges))
	for i, r := range ranges {
		ordinals[i] = ordered{Range: r, index: i}
	}

	sort.SliceStable(ordinals, func(i, j int) bool {
		return ordinals[i].Start < ordinals[j].Start
	})

	merged := []ordered{ordinals[0]}
	for _, cur := range ordinals[1:] {
		last := &merged[len(merged)-1]
		if cur.Start <= last.End+1 {
			// overlapping or adjacent: extend
			if cur.End > last.End {
				last.End = cur.End
			}
			if cur.index < last.index {
				last.index = cur.index
			}
		} else {
			merged = append(merged, cur)
		}
	}

	// Restore original ordering by the earliest index in each group.
	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].index < merged[j].index
	})

	out := make([]Range, len(merged))
	for i, m := range merged {
		out[i] = m.Range
	}
	return out
}
