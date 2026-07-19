package uniq

import (
	"math"
	"reflect"
	"testing"
)

// Upstream parity vectors for lodash's _.uniq and _.uniqBy.
//
// Source: lodash/lodash test suite, test/test.js (branch master), the shared
// block iterating ['uniq', 'uniqBy', 'uniqWith', 'sortedUniq', 'sortedUniqBy']
// beginning at the `QUnit.test('should return unique values of an unsorted
// array')` case, plus the `lodash.uniq` and `uniqBy methods` modules.
// Fetched from:
//   https://raw.githubusercontent.com/lodash/lodash/master/test/test.js
//
// Concrete lodash assertions ported here (values are lodash's, not invented):
//   uniq([2,1,2])                -> [2,1]        (unsorted)
//   uniq([1,2,2])                -> [1,2]        (sorted)
//   uniq treats object instances -> all kept     (reference identity)
//   uniq([-0, 0]) collapses      -> length 1     (-0 treated as 0)
//   uniq([NaN, NaN])             -> [NaN]        (SameValueZero matches NaN)
//   map(uniq, [[2,1,2],[1,2,1]]) -> [[2,1],[1,2]]
//   uniqBy(objects, o.a)         -> objects[0:3]
//   uniqBy(arrays, a[0])         -> arrays[0:3]
//   uniqBy([['a'],['a'],['b']])  -> [['a'],['b']]
//   uniqBy(repeated [1,2], key)  -> [[1,2]], first element preserved
//
// Divergence intentionally NOT asserted as lodash-equal (see notes): lodash
// treats two distinct object *references* with equal fields as unique, whereas
// Go struct values compare by value. Reference identity is modeled here with
// pointers, which does match lodash.

func TestParityUniqUnsorted(t *testing.T) {
	got := Uniq([]int{2, 1, 2})
	want := []int{2, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Uniq([2 1 2]) = %v, want %v", got, want)
	}
}

func TestParityUniqSorted(t *testing.T) {
	got := Uniq([]int{1, 2, 2})
	want := []int{1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Uniq([1 2 2]) = %v, want %v", got, want)
	}
}

// lodash: '`_.uniq` should treat object instances as unique'. In lodash each
// object literal is a distinct reference, so none are removed. Modeled with
// pointers, whose identity mirrors JS reference identity.
func TestParityUniqObjectInstancesUnique(t *testing.T) {
	type obj struct{ A int }
	objects := []*obj{{2}, {3}, {1}, {2}, {3}, {1}}
	got := Uniq(objects)
	if !reflect.DeepEqual(got, objects) {
		t.Errorf("Uniq(distinct pointers) removed elements: got len %d, want %d", len(got), len(objects))
	}
	// And a repeated identical pointer collapses.
	p := &obj{9}
	if g := Uniq([]*obj{p, p, p}); len(g) != 1 || g[0] != p {
		t.Errorf("Uniq([p p p]) = %v, want single p", g)
	}
}

// lodash: '`_.uniq` should treat `-0` as `0`'. func([-0, 0]) collapses to one.
func TestParityUniqNegativeZero(t *testing.T) {
	negZero := math.Copysign(0, -1)
	got := Uniq([]float64{negZero, 0})
	if len(got) != 1 {
		t.Errorf("Uniq([-0 0]) = %v, want length 1 (-0 treated as 0)", got)
	}
}

// lodash: '`_.uniq` should match `NaN`'. func([NaN, NaN]) -> [NaN].
func TestParityUniqMatchNaN(t *testing.T) {
	nan := math.NaN()
	got := Uniq([]float64{nan, nan})
	if len(got) != 1 || !math.IsNaN(got[0]) {
		t.Errorf("Uniq([NaN NaN]) = %v, want [NaN] (length 1)", got)
	}
	// A mix: distinct real values plus repeated NaNs.
	got2 := Uniq([]float64{1, nan, 2, nan, 1})
	if len(got2) != 3 || got2[0] != 1 || !math.IsNaN(got2[1]) || got2[2] != 2 {
		t.Errorf("Uniq([1 NaN 2 NaN 1]) = %v, want [1 NaN 2]", got2)
	}
}

// lodash.uniq module: 'should perform an unsorted uniq when used as an iteratee
// for methods like _.map'. map([[2,1,2],[1,2,1]], uniq) -> [[2,1],[1,2]].
func TestParityUniqAsIteratee(t *testing.T) {
	in := [][]int{{2, 1, 2}, {1, 2, 1}}
	got := make([][]int, len(in))
	for i, a := range in {
		got[i] = Uniq(a)
	}
	want := [][]int{{2, 1}, {1, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("map(Uniq) = %v, want %v", got, want)
	}
}

// uniqBy methods: '`_.uniqBy` should work with an `iteratee`'. Iteratee o.a over
// objects [{2},{3},{1},{2},{3},{1}] keeps the first three.
func TestParityUniqByIteratee(t *testing.T) {
	type obj struct{ A int }
	objects := []obj{{2}, {3}, {1}, {2}, {3}, {1}}
	got := UniqBy(objects, func(o obj) int { return o.A })
	want := objects[0:3]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UniqBy(objects, o.a) = %v, want %v", got, want)
	}
}

// uniqBy methods: '`_.uniqBy` should work with `_.property` shorthands' — the
// arrays branch: uniqBy([[2],[3],[1],[2],[3],[1]], index 0) keeps first three.
func TestParityUniqByArrayIndexKey(t *testing.T) {
	arrays := [][]int{{2}, {3}, {1}, {2}, {3}, {1}}
	got := UniqBy(arrays, func(a []int) int { return a[0] })
	want := arrays[0:3]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UniqBy(arrays, [0]) = %v, want %v", got, want)
	}
}

// uniqBy methods: '`_.uniqBy` should work with ... for `iteratee`' — the
// underlying data: func([['a'],['a'],['b']], keyOnFirst) -> [['a'],['b']].
func TestParityUniqByStringArrays(t *testing.T) {
	in := [][]string{{"a"}, {"a"}, {"b"}}
	got := UniqBy(in, func(a []string) string { return a[0] })
	want := [][]string{{"a"}, {"b"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UniqBy([['a']['a']['b']]) = %v, want %v", got, want)
	}
}

// uniqBy methods: 'should work with large arrays'. Every element is [1,2]; with
// a String-like key they all collapse to the first, and that first element is
// preserved by identity.
func TestParityUniqByLargeCollapse(t *testing.T) {
	const n = 200
	large := make([][]int, n)
	for i := range large {
		large[i] = []int{1, 2}
	}
	got := UniqBy(large, func(a []int) string {
		return "1,2" // stand-in for JS String([1,2]) == "1,2"
	})
	if len(got) != 1 {
		t.Fatalf("UniqBy(large) = len %d, want 1", len(got))
	}
	// lodash: assert.strictEqual(actual[0], largeArray[0]) — first preserved.
	if &got[0][0] != &large[0][0] {
		t.Errorf("UniqBy(large)[0] is not the first input element")
	}
	if !reflect.DeepEqual(got, [][]int{{1, 2}}) {
		t.Errorf("UniqBy(large) = %v, want [[1 2]]", got)
	}
}
