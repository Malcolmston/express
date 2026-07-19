package chunk

import (
	"reflect"
	"testing"
)

// Upstream parity vectors transcribed from lodash's own test suite:
//   https://raw.githubusercontent.com/lodash/lodash/master/test/test.js
//   QUnit.module('lodash.chunk') block (lines ~2536-2601).
//
// The shared fixture there is `var array = [0, 1, 2, 3, 4, 5]`. Each vector
// below reproduces a concrete input -> expected-output pair asserted upstream:
//
//   _.chunk(array, 3)                 => [[0,1,2],[3,4,5]]        (chunked arrays)
//   _.chunk(array, 4)                 => [[0,1,2,3],[4,5]]        (last chunk = remainder)
//   _.chunk(array, 0)                 => []                       (falsey size treated as 0)
//   _.chunk(array, -1)                => []                       (minimum size is 0)
//   _.chunk(array, array.length / 4)  => [[0],[1],..,[5]]         (size coerced/floored: 1.5 -> 1)
//   _.chunk([], 3)                    => []                       (empty input)
//
// Cases not representable in the typed Go port are intentionally omitted:
//   - the `undefined` size / default-arity case (_.chunk(array)) and the
//     `_.map(..., _.chunk)` iteratee-guard case both depend on lodash's dynamic
//     argument handling; Go's Chunk requires an explicit int size.

func TestParityChunkedArrays(t *testing.T) {
	got := Chunk([]int{0, 1, 2, 3, 4, 5}, 3)
	want := [][]int{{0, 1, 2}, {3, 4, 5}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Chunk(array,3) = %v, want %v", got, want)
	}
}

func TestParityLastChunkRemainder(t *testing.T) {
	got := Chunk([]int{0, 1, 2, 3, 4, 5}, 4)
	want := [][]int{{0, 1, 2, 3}, {4, 5}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Chunk(array,4) = %v, want %v", got, want)
	}
}

func TestParityFalseySizeZero(t *testing.T) {
	got := Chunk([]int{0, 1, 2, 3, 4, 5}, 0)
	want := [][]int{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Chunk(array,0) = %v, want %v", got, want)
	}
}

func TestParityMinimumSizeZero(t *testing.T) {
	got := Chunk([]int{0, 1, 2, 3, 4, 5}, -1)
	want := [][]int{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Chunk(array,-1) = %v, want %v", got, want)
	}
}

// array.length / 4 == 6 / 4 == 1.5, which lodash floors to 1.
func TestParitySizeFloored(t *testing.T) {
	got := Chunk([]int{0, 1, 2, 3, 4, 5}, 1)
	want := [][]int{{0}, {1}, {2}, {3}, {4}, {5}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Chunk(array,1) = %v, want %v", got, want)
	}
}

func TestParityEmptyInput(t *testing.T) {
	got := Chunk([]int{}, 3)
	want := [][]int{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Chunk([],3) = %v, want %v", got, want)
	}
}

// Additional canonical vector from the standard lodash docs / task spec:
//
//	_.chunk([0,1,2,3,4], 2) => [[0,1],[2,3],[4]]
func TestParityDocExample(t *testing.T) {
	got := Chunk([]int{0, 1, 2, 3, 4}, 2)
	want := [][]int{{0, 1}, {2, 3}, {4}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Chunk([0..4],2) = %v, want %v", got, want)
	}
}
