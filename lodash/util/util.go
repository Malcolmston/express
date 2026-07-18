// Package util ports helpers from lodash's "Util" category (cond, stubArray,
// stubTrue/False/String, toPath and a stepped range) to idiomatic, generic Go.
// These are the small building blocks lodash exposes for control flow and for
// providing constant "stub" values, offered here without any third-party
// dependency.
//
// Cond builds a function from an ordered list of predicate/transform pairs and
// runs the transform of the first matching predicate, the Go analogue of
// lodash.cond and a compact alternative to a switch when the branches are data.
// The Stub* helpers return fresh empty values (StubArray and StubObject each
// allocate a new, non-nil container) and the constant StubTrue/StubFalse/
// StubString mirror lodash's stubs for use as default callbacks. ToPath parses
// a lodash-style property path such as "a.b[0].c" into its string segments,
// and RangeStep generates an arithmetic sequence with an explicit step, filling
// the gap left by the plain Range helpers.
//
// Everything is deterministic and depends only on the standard library.
package util

import (
	"strconv"
	"strings"
)

// CondPair is a single predicate/transform branch for Cond: When selects the
// branch and Then produces its result.
type CondPair[T, R any] struct {
	When func(T) bool
	Then func(T) R
}

// Cond returns a function that evaluates each pair's predicate in order and,
// for the first whose predicate is satisfied, returns that pair's transform
// applied to the argument together with true. If no predicate matches it
// returns the zero value of R and false. It mirrors lodash.cond.
func Cond[T, R any](pairs ...CondPair[T, R]) func(T) (R, bool) {
	return func(v T) (R, bool) {
		for _, p := range pairs {
			if p.When(v) {
				return p.Then(v), true
			}
		}
		var zero R
		return zero, false
	}
}

// StubArray returns a new empty, non-nil slice of T, mirroring lodash.stubArray.
func StubArray[T any]() []T { return []T{} }

// StubObject returns a new empty, non-nil map, mirroring lodash.stubObject.
func StubObject[K comparable, V any]() map[K]V { return map[K]V{} }

// StubTrue always returns true, mirroring lodash.stubTrue.
func StubTrue() bool { return true }

// StubFalse always returns false, mirroring lodash.stubFalse.
func StubFalse() bool { return false }

// StubString always returns the empty string, mirroring lodash.stubString.
func StubString() string { return "" }

// ToPath splits a lodash-style property path into its segments. It understands
// dot notation and bracket indexing with optional quotes, so ToPath("a.b[0].c")
// returns ["a", "b", "0", "c"] and ToPath(`a["x.y"]`) returns ["a", "x.y"].
func ToPath(path string) []string {
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	i := 0
	for i < len(path) {
		c := path[i]
		switch c {
		case '.':
			flush()
			i++
		case '[':
			flush()
			i++
			if i < len(path) && (path[i] == '"' || path[i] == '\'') {
				quote := path[i]
				i++
				start := i
				for i < len(path) && path[i] != quote {
					i++
				}
				out = append(out, path[start:i])
				i++ // closing quote
				if i < len(path) && path[i] == ']' {
					i++
				}
			} else {
				start := i
				for i < len(path) && path[i] != ']' {
					i++
				}
				out = append(out, path[start:i])
				if i < len(path) {
					i++ // closing bracket
				}
			}
		default:
			cur.WriteByte(c)
			i++
		}
	}
	flush()
	return out
}

// RangeStep returns the arithmetic sequence starting at start, advancing by
// step, up to but not including end. A positive step produces an ascending
// sequence and a negative step a descending one; a zero step, or a step whose
// sign cannot reach end, yields an empty slice. It mirrors lodash.range's
// three-argument form for int.
func RangeStep(start, end, step int) []int {
	out := []int{}
	if step == 0 {
		return out
	}
	if step > 0 {
		for v := start; v < end; v += step {
			out = append(out, v)
		}
	} else {
		for v := start; v > end; v += step {
			out = append(out, v)
		}
	}
	return out
}

// ToIndex parses a single path segment as a non-negative array index, returning
// the index and true when the segment is a valid unsigned integer literal.
func ToIndex(segment string) (int, bool) {
	n, err := strconv.Atoi(segment)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}
