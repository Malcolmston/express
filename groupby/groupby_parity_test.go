package groupby

import (
	"reflect"
	"testing"
)

// Parity vectors transcribed from the canonical lodash test suite:
//   https://raw.githubusercontent.com/lodash/lodash/main/test/test.js
//   QUnit.module('lodash.groupBy') at lines ~7697-7758.
//
// lodash coerces every computed key to a string property name on a plain
// object; this Go port keeps the key's native type as the map key. The vectors
// below use native Go key types (int/string) that correspond one-to-one with
// lodash's stringified keys, and iteratee shorthands ('length', property
// index 0/1, nullish->identity) are expressed as equivalent Go key functions.

// TestParityFloorIteratee mirrors:
//
//	_.groupBy([6.1, 4.2, 6.3], Math.floor)
//	=> { '4': [4.2], '6': [6.1, 6.3] }
func TestParityFloorIteratee(t *testing.T) {
	in := []float64{6.1, 4.2, 6.3}
	got := GroupBy(in, func(f float64) int { return int(f) }) // Math.floor for positives
	want := map[int][]float64{
		4: {4.2},
		6: {6.1, 6.3},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v, want %v", got, want)
	}
}

// TestParityLengthShorthand mirrors:
//
//	_.groupBy(['one', 'two', 'three'], 'length')
//	=> { '3': ['one', 'two'], '5': ['three'] }
func TestParityLengthShorthand(t *testing.T) {
	in := []string{"one", "two", "three"}
	got := GroupBy(in, func(s string) int { return len(s) })
	want := map[int][]string{
		3: {"one", "two"},
		5: {"three"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v, want %v", got, want)
	}
}

// pair models a lodash [number, string] tuple used by the property-index tests.
type pair struct {
	N int
	S string
}

// TestParityIndexZeroShorthand mirrors:
//
//	_.groupBy([[1,'a'],[2,'a'],[2,'b']], 0)
//	=> { '1': [[1,'a']], '2': [[2,'a'],[2,'b']] }
func TestParityIndexZeroShorthand(t *testing.T) {
	in := []pair{{1, "a"}, {2, "a"}, {2, "b"}}
	got := GroupBy(in, func(p pair) int { return p.N })
	want := map[int][]pair{
		1: {{1, "a"}},
		2: {{2, "a"}, {2, "b"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v, want %v", got, want)
	}
}

// TestParityIndexOneShorthand mirrors:
//
//	_.groupBy([[1,'a'],[2,'a'],[2,'b']], 1)
//	=> { 'a': [[1,'a'],[2,'a']], 'b': [[2,'b']] }
func TestParityIndexOneShorthand(t *testing.T) {
	in := []pair{{1, "a"}, {2, "a"}, {2, "b"}}
	got := GroupBy(in, func(p pair) string { return p.S })
	want := map[string][]pair{
		"a": {{1, "a"}, {2, "a"}},
		"b": {{2, "b"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v, want %v", got, want)
	}
}

// TestParityIdentityNullish mirrors:
//
//	_.groupBy([6, 4, 6])  // nullish iteratee -> _.identity
//	=> { '4': [4], '6': [6, 6] }
func TestParityIdentityNullish(t *testing.T) {
	in := []int{6, 4, 6}
	got := GroupBy(in, func(n int) int { return n })
	want := map[int][]int{
		4: {4},
		6: {6, 6},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v, want %v", got, want)
	}
}

// TestParityReservedKeyNames mirrors:
//
//	_.groupBy([6.1, 4.2, 6.3], n => Math.floor(n) > 4 ? 'hasOwnProperty' : 'constructor')
//	=> { constructor: [4.2], hasOwnProperty: [6.1, 6.3] }
//
// In lodash this exercises own-vs-inherited property safety; a Go map keys on
// plain strings so the reserved-name keys behave like any other key.
func TestParityReservedKeyNames(t *testing.T) {
	in := []float64{6.1, 4.2, 6.3}
	got := GroupBy(in, func(n float64) string {
		if int(n) > 4 {
			return "hasOwnProperty"
		}
		return "constructor"
	})
	want := map[string][]float64{
		"constructor":    {4.2},
		"hasOwnProperty": {6.1, 6.3},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v, want %v", got, want)
	}
}
