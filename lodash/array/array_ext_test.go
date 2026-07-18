package array

import (
	"reflect"
	"testing"
)

func TestDifferenceWith(t *testing.T) {
	eq := func(a, b int) bool { return a == b }
	got := DifferenceWith([]int{1, 2, 3, 4}, []int{2, 4}, eq)
	if !reflect.DeepEqual(got, []int{1, 3}) {
		t.Errorf("DifferenceWith = %v", got)
	}
}

func TestIntersectionWith(t *testing.T) {
	eq := func(a, b int) bool { return a == b }
	got := IntersectionWith([]int{1, 2, 2, 3}, []int{2, 3, 4}, eq)
	if !reflect.DeepEqual(got, []int{2, 3}) {
		t.Errorf("IntersectionWith = %v", got)
	}
}

func TestUnionWith(t *testing.T) {
	eq := func(a, b int) bool { return a == b }
	got := UnionWith(eq, []int{1, 2}, []int{2, 3}, []int{3, 4})
	if !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Errorf("UnionWith = %v", got)
	}
}

func TestXorByWith(t *testing.T) {
	got := XorBy(func(n int) int { return n }, []int{1, 2, 3}, []int{2, 3, 4})
	if !reflect.DeepEqual(got, []int{1, 4}) {
		t.Errorf("XorBy = %v", got)
	}
	eq := func(a, b int) bool { return a == b }
	got2 := XorWith(eq, []int{2, 1}, []int{2, 3})
	if !reflect.DeepEqual(got2, []int{1, 3}) {
		t.Errorf("XorWith = %v", got2)
	}
}

func TestZipUnzipWith(t *testing.T) {
	sum := func(g []int) int {
		s := 0
		for _, v := range g {
			s += v
		}
		return s
	}
	got := ZipWith(sum, []int{1, 2}, []int{10, 20}, []int{100, 200})
	if !reflect.DeepEqual(got, []int{111, 222}) {
		t.Errorf("ZipWith = %v", got)
	}
	got2 := UnzipWith(sum, [][]int{{1, 10, 100}, {2, 20, 200}})
	if !reflect.DeepEqual(got2, []int{3, 30, 300}) {
		t.Errorf("UnzipWith = %v", got2)
	}
}

func TestPullAt(t *testing.T) {
	in := []int{10, 20, 30, 40, 50}
	got := PullAt(in, 1, 3, 9)
	if !reflect.DeepEqual(got, []int{10, 30, 50}) {
		t.Errorf("PullAt = %v", got)
	}
	if !reflect.DeepEqual(in, []int{10, 20, 30, 40, 50}) {
		t.Errorf("PullAt mutated input: %v", in)
	}
}

func TestSortedSearch(t *testing.T) {
	s := []int{1, 2, 2, 2, 3, 5}
	if got := SortedIndexOf(s, 2); got != 1 {
		t.Errorf("SortedIndexOf = %d, want 1", got)
	}
	if got := SortedIndexOf(s, 4); got != -1 {
		t.Errorf("SortedIndexOf missing = %d", got)
	}
	if got := SortedLastIndexOf(s, 2); got != 3 {
		t.Errorf("SortedLastIndexOf = %d, want 3", got)
	}
	if got := SortedLastIndex(s, 2); got != 4 {
		t.Errorf("SortedLastIndex = %d, want 4", got)
	}
	if got := SortedIndexBy(s, 4, func(n int) int { return n }); got != 5 {
		t.Errorf("SortedIndexBy = %d, want 5", got)
	}
}

func TestSortedUniq(t *testing.T) {
	if got := SortedUniq([]int{1, 1, 2, 3, 3, 3, 4}); !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Errorf("SortedUniq = %v", got)
	}
	got := SortedUniqBy([]float64{1.1, 1.9, 2.2, 2.8}, func(f float64) int { return int(f) })
	if !reflect.DeepEqual(got, []float64{1.1, 2.2}) {
		t.Errorf("SortedUniqBy = %v", got)
	}
}

func TestRightWhile(t *testing.T) {
	pred := func(n int) bool { return n > 2 }
	if got := TakeRightWhile([]int{1, 2, 3, 4}, pred); !reflect.DeepEqual(got, []int{3, 4}) {
		t.Errorf("TakeRightWhile = %v", got)
	}
	if got := DropRightWhile([]int{1, 2, 3, 4}, pred); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("DropRightWhile = %v", got)
	}
}

func BenchmarkSortedIndexOf(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for i := 0; i < b.N; i++ {
		_ = SortedIndexOf(s, 512)
	}
}
