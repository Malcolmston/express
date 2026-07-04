package array

import (
	"math"
	"reflect"
	"testing"
)

func floorF(v float64) float64 { return math.Floor(v) }

func TestCompact(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"mixed", []int{0, 1, 0, 2, 3}, []int{1, 2, 3}},
		{"empty", []int{}, []int{}},
		{"all zero", []int{0, 0}, []int{}},
		{"no zero", []int{1, 2}, []int{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compact(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Compact(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestChunk(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		size int
		want [][]string
	}{
		{"size2", []string{"a", "b", "c", "d"}, 2, [][]string{{"a", "b"}, {"c", "d"}}},
		{"size3", []string{"a", "b", "c", "d"}, 3, [][]string{{"a", "b", "c"}, {"d"}}},
		{"size0", []string{"a", "b"}, 0, [][]string{}},
		{"negative", []string{"a", "b"}, -1, [][]string{}},
		{"empty", []string{}, 2, [][]string{}},
		{"bigger than len", []string{"a"}, 5, [][]string{{"a"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Chunk(tt.in, tt.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Chunk(%v, %d) = %v, want %v", tt.in, tt.size, got, tt.want)
			}
		})
	}
}

func TestDifference(t *testing.T) {
	if got := Difference([]int{2, 1}, []int{2, 3}); !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("Difference = %v, want [1]", got)
	}
	if got := Difference([]int{1, 2, 3}, []int{2}, []int{3}); !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("Difference multi = %v, want [1]", got)
	}
	if got := Difference([]int{1, 2}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("Difference no others = %v, want [1 2]", got)
	}
}

func TestDifferenceBy(t *testing.T) {
	got := DifferenceBy([]float64{2.1, 1.2}, floorF, []float64{2.3, 3.4})
	if !reflect.DeepEqual(got, []float64{1.2}) {
		t.Errorf("DifferenceBy = %v, want [1.2]", got)
	}
}

func TestIntersection(t *testing.T) {
	if got := Intersection([]int{2, 1}, []int{2, 3}); !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("Intersection = %v, want [2]", got)
	}
	if got := Intersection([]int{1, 1, 2}, []int{1, 2, 2}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("Intersection dedupe = %v, want [1 2]", got)
	}
	if got := Intersection[int](); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("Intersection empty = %v, want []", got)
	}
}

func TestIntersectionBy(t *testing.T) {
	got := IntersectionBy(floorF, []float64{2.1, 1.2}, []float64{2.3, 3.4})
	if !reflect.DeepEqual(got, []float64{2.1}) {
		t.Errorf("IntersectionBy = %v, want [2.1]", got)
	}
}

func TestUnion(t *testing.T) {
	if got := Union([]int{2}, []int{1, 2}); !reflect.DeepEqual(got, []int{2, 1}) {
		t.Errorf("Union = %v, want [2 1]", got)
	}
	if got := Union([]int{}, []int{}); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("Union empty = %v, want []", got)
	}
}

func TestUnionBy(t *testing.T) {
	got := UnionBy(floorF, []float64{2.1}, []float64{1.2, 2.3})
	if !reflect.DeepEqual(got, []float64{2.1, 1.2}) {
		t.Errorf("UnionBy = %v, want [2.1 1.2]", got)
	}
}

func TestWithout(t *testing.T) {
	if got := Without([]int{2, 1, 2, 3}, 1, 2); !reflect.DeepEqual(got, []int{3}) {
		t.Errorf("Without = %v, want [3]", got)
	}
}

func TestXor(t *testing.T) {
	if got := Xor([]int{2, 1}, []int{2, 3}); !reflect.DeepEqual(got, []int{1, 3}) {
		t.Errorf("Xor = %v, want [1 3]", got)
	}
	if got := Xor([]int{1, 2}, []int{1, 2}); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("Xor same = %v, want []", got)
	}
}

func TestUniq(t *testing.T) {
	if got := Uniq([]int{2, 1, 2}); !reflect.DeepEqual(got, []int{2, 1}) {
		t.Errorf("Uniq = %v, want [2 1]", got)
	}
}

func TestUniqBy(t *testing.T) {
	got := UniqBy([]float64{2.1, 1.2, 2.3}, floorF)
	if !reflect.DeepEqual(got, []float64{2.1, 1.2}) {
		t.Errorf("UniqBy = %v, want [2.1 1.2]", got)
	}
}

func TestDropTake(t *testing.T) {
	base := []int{1, 2, 3}
	if got := Drop(base, 2); !reflect.DeepEqual(got, []int{3}) {
		t.Errorf("Drop = %v, want [3]", got)
	}
	if got := Drop(base, 5); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("Drop clamp = %v, want []", got)
	}
	if got := DropRight(base, 2); !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("DropRight = %v, want [1]", got)
	}
	if got := Take(base, 2); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("Take = %v, want [1 2]", got)
	}
	if got := Take(base, 0); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("Take 0 = %v, want []", got)
	}
	if got := TakeRight(base, 2); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Errorf("TakeRight = %v, want [2 3]", got)
	}
	if got := Take(base, 10); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("Take clamp = %v, want [1 2 3]", got)
	}
}

func TestDropTakeWhile(t *testing.T) {
	in := []int{1, 2, 3, 4}
	lt3 := func(v int) bool { return v < 3 }
	if got := DropWhile(in, lt3); !reflect.DeepEqual(got, []int{3, 4}) {
		t.Errorf("DropWhile = %v, want [3 4]", got)
	}
	if got := TakeWhile(in, lt3); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("TakeWhile = %v, want [1 2]", got)
	}
}

func TestHeadLast(t *testing.T) {
	if v, ok := Head([]int{1, 2, 3}); v != 1 || !ok {
		t.Errorf("Head = %v,%v want 1,true", v, ok)
	}
	if v, ok := Head([]int{}); v != 0 || ok {
		t.Errorf("Head empty = %v,%v want 0,false", v, ok)
	}
	if v, ok := Last([]int{1, 2, 3}); v != 3 || !ok {
		t.Errorf("Last = %v,%v want 3,true", v, ok)
	}
	if _, ok := Last([]int{}); ok {
		t.Errorf("Last empty ok = true, want false")
	}
}

func TestTailInitial(t *testing.T) {
	if got := Tail([]int{1, 2, 3}); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Errorf("Tail = %v, want [2 3]", got)
	}
	if got := Tail([]int{}); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("Tail empty = %v, want []", got)
	}
	if got := Initial([]int{1, 2, 3}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("Initial = %v, want [1 2]", got)
	}
	if got := Initial([]int{}); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("Initial empty = %v, want []", got)
	}
}

func TestNth(t *testing.T) {
	in := []int{1, 2, 3}
	if v, ok := Nth(in, 1); v != 2 || !ok {
		t.Errorf("Nth 1 = %v,%v want 2,true", v, ok)
	}
	if v, ok := Nth(in, -1); v != 3 || !ok {
		t.Errorf("Nth -1 = %v,%v want 3,true", v, ok)
	}
	if _, ok := Nth(in, 5); ok {
		t.Errorf("Nth oob = true, want false")
	}
	if _, ok := Nth(in, -5); ok {
		t.Errorf("Nth neg oob = true, want false")
	}
}

func TestFindIndex(t *testing.T) {
	in := []int{1, 2, 3}
	if got := FindIndex(in, func(v int) bool { return v == 2 }); got != 1 {
		t.Errorf("FindIndex = %d, want 1", got)
	}
	if got := FindIndex(in, func(v int) bool { return v == 9 }); got != -1 {
		t.Errorf("FindIndex miss = %d, want -1", got)
	}
	if got := FindLastIndex([]int{1, 2, 1}, func(v int) bool { return v == 1 }); got != 2 {
		t.Errorf("FindLastIndex = %d, want 2", got)
	}
}

func TestIndexOf(t *testing.T) {
	in := []int{1, 2, 1, 2}
	if got := IndexOf(in, 2); got != 1 {
		t.Errorf("IndexOf = %d, want 1", got)
	}
	if got := IndexOf(in, 9); got != -1 {
		t.Errorf("IndexOf miss = %d, want -1", got)
	}
	if got := LastIndexOf(in, 2); got != 3 {
		t.Errorf("LastIndexOf = %d, want 3", got)
	}
	if got := LastIndexOf(in, 9); got != -1 {
		t.Errorf("LastIndexOf miss = %d, want -1", got)
	}
}

func TestFill(t *testing.T) {
	in := []int{1, 2, 3, 4}
	if got := Fill(in, 0, 1, 3); !reflect.DeepEqual(got, []int{1, 0, 0, 4}) {
		t.Errorf("Fill = %v, want [1 0 0 4]", got)
	}
	if got := Fill(in, 9, -2, 4); !reflect.DeepEqual(got, []int{1, 2, 9, 9}) {
		t.Errorf("Fill neg = %v, want [1 2 9 9]", got)
	}
	// input must be unchanged
	if !reflect.DeepEqual(in, []int{1, 2, 3, 4}) {
		t.Errorf("Fill mutated input: %v", in)
	}
}

func TestFlatten(t *testing.T) {
	if got := Flatten([][]int{{1, 2}, {3, 4}}); !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Errorf("Flatten = %v, want [1 2 3 4]", got)
	}
	if got := Flatten([][]int{}); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("Flatten empty = %v, want []", got)
	}
}

func TestFlattenDeep(t *testing.T) {
	in := []any{1, []any{2, []any{3, 4}}, 5}
	got := FlattenDeep(in)
	want := []any{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FlattenDeep = %v, want %v", got, want)
	}
}

func TestFromPairs(t *testing.T) {
	got := FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}})
	want := map[string]int{"a": 1, "b": 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FromPairs = %v, want %v", got, want)
	}
}

func TestZipUnzip(t *testing.T) {
	got := Zip([]int{1, 2}, []int{3, 4})
	want := [][]int{{1, 3}, {2, 4}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Zip = %v, want %v", got, want)
	}
	// uneven lengths -> zero fill
	uneven := Zip([]int{1, 2, 3}, []int{4})
	wantU := [][]int{{1, 4}, {2, 0}, {3, 0}}
	if !reflect.DeepEqual(uneven, wantU) {
		t.Errorf("Zip uneven = %v, want %v", uneven, wantU)
	}
	back := Unzip(got)
	if !reflect.DeepEqual(back, [][]int{{1, 2}, {3, 4}}) {
		t.Errorf("Unzip = %v, want [[1 2] [3 4]]", back)
	}
}

func TestZipObject(t *testing.T) {
	got := ZipObject([]string{"a", "b"}, []int{1, 2})
	want := map[string]int{"a": 1, "b": 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ZipObject = %v, want %v", got, want)
	}
	// missing value -> zero
	got2 := ZipObject([]string{"a", "b"}, []int{1})
	want2 := map[string]int{"a": 1, "b": 0}
	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("ZipObject short = %v, want %v", got2, want2)
	}
}

func TestReverse(t *testing.T) {
	in := []int{1, 2, 3}
	if got := Reverse(in); !reflect.DeepEqual(got, []int{3, 2, 1}) {
		t.Errorf("Reverse = %v, want [3 2 1]", got)
	}
	if !reflect.DeepEqual(in, []int{1, 2, 3}) {
		t.Errorf("Reverse mutated input: %v", in)
	}
}

func TestConcat(t *testing.T) {
	got := Concat([]int{1}, []int{2, 3}, []int{4})
	if !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Errorf("Concat = %v, want [1 2 3 4]", got)
	}
	if got := Concat[int](); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("Concat empty = %v, want []", got)
	}
}

func TestPullRemove(t *testing.T) {
	in := []int{1, 2, 3, 1, 2}
	if got := Pull(in, 2, 3); !reflect.DeepEqual(got, []int{1, 1}) {
		t.Errorf("Pull = %v, want [1 1]", got)
	}
	if got := PullAll(in, []int{2, 3}); !reflect.DeepEqual(got, []int{1, 1}) {
		t.Errorf("PullAll = %v, want [1 1]", got)
	}
	if !reflect.DeepEqual(in, []int{1, 2, 3, 1, 2}) {
		t.Errorf("Pull mutated input: %v", in)
	}
	got := Remove([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 })
	if !reflect.DeepEqual(got, []int{1, 3}) {
		t.Errorf("Remove = %v, want [1 3]", got)
	}
}

func TestSortedIndex(t *testing.T) {
	tests := []struct {
		in    []int
		value int
		want  int
	}{
		{[]int{30, 50}, 40, 1},
		{[]int{20, 30, 50}, 10, 0},
		{[]int{20, 30, 50}, 60, 3},
		{[]int{20, 30, 30, 50}, 30, 1},
		{[]int{}, 5, 0},
	}
	for _, tt := range tests {
		if got := SortedIndex(tt.in, tt.value); got != tt.want {
			t.Errorf("SortedIndex(%v, %d) = %d, want %d", tt.in, tt.value, got, tt.want)
		}
	}
}

func TestSlice(t *testing.T) {
	in := []int{1, 2, 3, 4}
	if got := Slice(in, 1, 3); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Errorf("Slice = %v, want [2 3]", got)
	}
	if got := Slice(in, -2, 4); !reflect.DeepEqual(got, []int{3, 4}) {
		t.Errorf("Slice neg = %v, want [3 4]", got)
	}
	if got := Slice(in, 3, 1); !reflect.DeepEqual(got, []int{}) {
		t.Errorf("Slice inverted = %v, want []", got)
	}
	if got := Slice(in, 0, 10); !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Errorf("Slice clamp = %v, want [1 2 3 4]", got)
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		name             string
		start, end, step int
		want             []int
	}{
		{"basic", 0, 4, 1, []int{0, 1, 2, 3}},
		{"step2", 0, 6, 2, []int{0, 2, 4}},
		{"down", 4, 0, -1, []int{4, 3, 2, 1}},
		{"default step down", 4, 0, 0, []int{4, 3, 2, 1}},
		{"default step up", 0, 3, 0, []int{0, 1, 2}},
		{"empty", 0, 0, 1, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Range(tt.start, tt.end, tt.step); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Range(%d,%d,%d) = %v, want %v", tt.start, tt.end, tt.step, got, tt.want)
			}
		})
	}
}

func TestRangeRight(t *testing.T) {
	if got := RangeRight(0, 4, 1); !reflect.DeepEqual(got, []int{3, 2, 1, 0}) {
		t.Errorf("RangeRight = %v, want [3 2 1 0]", got)
	}
}
