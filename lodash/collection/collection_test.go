package collection

import (
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type person struct {
	Name string
	Age  int
}

func TestCountBy(t *testing.T) {
	got := CountBy([]float64{6.1, 4.2, 6.3}, func(f float64) int { return int(f) })
	want := map[int]int{6: 2, 4: 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CountBy = %v, want %v", got, want)
	}
	if len(CountBy([]int{}, func(i int) int { return i })) != 0 {
		t.Fatalf("CountBy empty should be empty")
	}
}

func TestKeyBy(t *testing.T) {
	people := []person{{"a", 1}, {"b", 2}, {"a", 3}}
	got := KeyBy(people, func(p person) string { return p.Name })
	if got["a"].Age != 3 {
		t.Fatalf("KeyBy last-wins failed: %+v", got["a"])
	}
	if got["b"].Age != 2 {
		t.Fatalf("KeyBy b failed: %+v", got["b"])
	}
}

func TestGroupBy(t *testing.T) {
	got := GroupBy([]float64{6.1, 4.2, 6.3}, func(f float64) int { return int(f) })
	want := map[int][]float64{6: {6.1, 6.3}, 4: {4.2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GroupBy = %v, want %v", got, want)
	}
}

func TestPartition(t *testing.T) {
	truthy, falsy := Partition([]int{1, 2, 3, 4, 5}, func(i int) bool { return i%2 == 0 })
	if !reflect.DeepEqual(truthy, []int{2, 4}) {
		t.Fatalf("Partition truthy = %v", truthy)
	}
	if !reflect.DeepEqual(falsy, []int{1, 3, 5}) {
		t.Fatalf("Partition falsy = %v", falsy)
	}
}

func TestEverySome(t *testing.T) {
	nums := []int{2, 4, 6}
	if !Every(nums, func(i int) bool { return i%2 == 0 }) {
		t.Fatal("Every should be true")
	}
	if Every(nums, func(i int) bool { return i > 4 }) {
		t.Fatal("Every should be false")
	}
	if !Every([]int{}, func(i int) bool { return false }) {
		t.Fatal("Every on empty should be true (vacuous)")
	}
	if !Some(nums, func(i int) bool { return i > 4 }) {
		t.Fatal("Some should be true")
	}
	if Some(nums, func(i int) bool { return i > 10 }) {
		t.Fatal("Some should be false")
	}
	if Some([]int{}, func(i int) bool { return true }) {
		t.Fatal("Some on empty should be false")
	}
}

func TestFilterReject(t *testing.T) {
	nums := []int{1, 2, 3, 4}
	even := func(i int) bool { return i%2 == 0 }
	if !reflect.DeepEqual(Filter(nums, even), []int{2, 4}) {
		t.Fatalf("Filter failed: %v", Filter(nums, even))
	}
	if !reflect.DeepEqual(Reject(nums, even), []int{1, 3}) {
		t.Fatalf("Reject failed: %v", Reject(nums, even))
	}
}

func TestFindFindLast(t *testing.T) {
	nums := []int{1, 2, 3, 4}
	if v, ok := Find(nums, func(i int) bool { return i%2 == 0 }); !ok || v != 2 {
		t.Fatalf("Find = %v,%v", v, ok)
	}
	if v, ok := FindLast(nums, func(i int) bool { return i%2 == 0 }); !ok || v != 4 {
		t.Fatalf("FindLast = %v,%v", v, ok)
	}
	if _, ok := Find(nums, func(i int) bool { return i > 10 }); ok {
		t.Fatal("Find should not find")
	}
	if _, ok := FindLast(nums, func(i int) bool { return i > 10 }); ok {
		t.Fatal("FindLast should not find")
	}
}

func TestForEachEachEarlyStop(t *testing.T) {
	var seen []int
	ForEach([]int{1, 2, 3, 4}, func(i int) bool {
		seen = append(seen, i)
		return i < 2 // stop after 2
	})
	if !reflect.DeepEqual(seen, []int{1, 2}) {
		t.Fatalf("ForEach early stop = %v", seen)
	}
	var all []int
	Each([]int{1, 2, 3}, func(i int) bool { all = append(all, i); return true })
	if !reflect.DeepEqual(all, []int{1, 2, 3}) {
		t.Fatalf("Each = %v", all)
	}
}

func TestMapFlatMap(t *testing.T) {
	got := Map([]int{1, 2, 3}, func(i int) int { return i * i })
	if !reflect.DeepEqual(got, []int{1, 4, 9}) {
		t.Fatalf("Map = %v", got)
	}
	strs := Map([]int{1, 2}, func(i int) string { return strings.Repeat("x", i) })
	if !reflect.DeepEqual(strs, []string{"x", "xx"}) {
		t.Fatalf("Map to string = %v", strs)
	}
	fm := FlatMap([]int{1, 2}, func(i int) []int { return []int{i, i} })
	if !reflect.DeepEqual(fm, []int{1, 1, 2, 2}) {
		t.Fatalf("FlatMap = %v", fm)
	}
}

func TestReduceReduceRight(t *testing.T) {
	sum := Reduce([]int{1, 2, 3, 4}, func(acc, cur int) int { return acc + cur }, 0)
	if sum != 10 {
		t.Fatalf("Reduce sum = %d", sum)
	}
	l := Reduce([]string{"a", "b", "c"}, func(acc, cur string) string { return acc + cur }, "")
	if l != "abc" {
		t.Fatalf("Reduce concat = %q", l)
	}
	r := ReduceRight([]string{"a", "b", "c"}, func(acc, cur string) string { return acc + cur }, "")
	if r != "cba" {
		t.Fatalf("ReduceRight = %q", r)
	}
}

func TestIncludesSize(t *testing.T) {
	if !Includes([]int{1, 2, 3}, 2) {
		t.Fatal("Includes should be true")
	}
	if Includes([]string{"a", "b"}, "c") {
		t.Fatal("Includes should be false")
	}
	if Size([]int{1, 2, 3, 4}) != 4 {
		t.Fatal("Size wrong")
	}
}

func TestSortBy(t *testing.T) {
	people := []person{{"c", 3}, {"a", 1}, {"b", 2}}
	got := SortBy(people, func(p person) string { return p.Name })
	names := Map(got, func(p person) string { return p.Name })
	if !reflect.DeepEqual(names, []string{"a", "b", "c"}) {
		t.Fatalf("SortBy = %v", names)
	}
	// original untouched
	if people[0].Name != "c" {
		t.Fatal("SortBy mutated input")
	}
	// stability: equal keys preserve order
	pairs := []person{{"x", 2}, {"y", 1}, {"z", 2}, {"w", 1}}
	byAge := SortBy(pairs, func(p person) int { return p.Age })
	got2 := Map(byAge, func(p person) string { return p.Name })
	if !reflect.DeepEqual(got2, []string{"y", "w", "x", "z"}) {
		t.Fatalf("SortBy stability = %v", got2)
	}
}

func TestOrderBy(t *testing.T) {
	people := []person{
		{"barney", 36},
		{"fred", 40},
		{"barney", 34},
		{"fred", 48},
	}
	got := OrderBy(people,
		[]func(person) any{
			func(p person) any { return p.Name },
			func(p person) any { return p.Age },
		},
		[]string{"asc", "desc"},
	)
	want := []person{
		{"barney", 36},
		{"barney", 34},
		{"fred", 48},
		{"fred", 40},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("OrderBy = %v, want %v", got, want)
	}
	// input untouched
	if people[0].Age != 36 {
		t.Fatal("OrderBy mutated input")
	}
}

func TestMinMaxSumMean(t *testing.T) {
	people := []person{{"a", 30}, {"b", 10}, {"c", 20}}
	if mn, ok := MinBy(people, func(p person) int { return p.Age }); !ok || mn.Name != "b" {
		t.Fatalf("MinBy = %+v %v", mn, ok)
	}
	if mx, ok := MaxBy(people, func(p person) int { return p.Age }); !ok || mx.Name != "a" {
		t.Fatalf("MaxBy = %+v %v", mx, ok)
	}
	if s := SumBy(people, func(p person) int { return p.Age }); s != 60 {
		t.Fatalf("SumBy = %d", s)
	}
	if m := MeanBy(people, func(p person) int { return p.Age }); m != 20 {
		t.Fatalf("MeanBy = %v", m)
	}
	if _, ok := MinBy([]person{}, func(p person) int { return p.Age }); ok {
		t.Fatal("MinBy empty should be false")
	}
	if MeanBy([]person{}, func(p person) int { return p.Age }) != 0 {
		t.Fatal("MeanBy empty should be 0")
	}
	if SumBy([]float64{1.5, 2.5}, func(f float64) float64 { return f }) != 4.0 {
		t.Fatal("SumBy float wrong")
	}
}

func TestShuffleIsPermutation(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	in := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	out := Shuffle(in, r)
	if len(out) != len(in) {
		t.Fatalf("Shuffle length changed: %d", len(out))
	}
	// input unchanged
	if !reflect.DeepEqual(in, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
		t.Fatal("Shuffle mutated input")
	}
	// permutation: same multiset
	a := append([]int(nil), out...)
	b := append([]int(nil), in...)
	sort.Ints(a)
	sort.Ints(b)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Shuffle not a permutation: %v", out)
	}
	// deterministic given seed
	out2 := Shuffle(in, rand.New(rand.NewSource(42)))
	if !reflect.DeepEqual(out, out2) {
		t.Fatalf("Shuffle not deterministic: %v vs %v", out, out2)
	}
}

func TestSample(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	in := []string{"a", "b", "c"}
	v, ok := Sample(in, r)
	if !ok || !Includes(in, v) {
		t.Fatalf("Sample = %v, %v", v, ok)
	}
	if _, ok := Sample([]int{}, r); ok {
		t.Fatal("Sample empty should be false")
	}
}

func TestSampleSize(t *testing.T) {
	r := rand.New(rand.NewSource(99))
	in := []int{1, 2, 3, 4, 5}
	got := SampleSize(in, 3, r)
	if len(got) != 3 {
		t.Fatalf("SampleSize count = %d", len(got))
	}
	// without replacement: all distinct and subset of in
	seen := map[int]bool{}
	for _, v := range got {
		if seen[v] {
			t.Fatalf("SampleSize duplicate: %v", got)
		}
		if !Includes(in, v) {
			t.Fatalf("SampleSize foreign element: %v", v)
		}
		seen[v] = true
	}
	// n >= len returns full permutation
	full := SampleSize(in, 100, r)
	if len(full) != len(in) {
		t.Fatalf("SampleSize overflow len = %d", len(full))
	}
	fa := append([]int(nil), full...)
	fb := append([]int(nil), in...)
	sort.Ints(fa)
	sort.Ints(fb)
	if !reflect.DeepEqual(fa, fb) {
		t.Fatalf("SampleSize full not permutation: %v", full)
	}
	if len(SampleSize(in, 0, r)) != 0 {
		t.Fatal("SampleSize 0 should be empty")
	}
	if len(SampleSize(in, -1, r)) != 0 {
		t.Fatal("SampleSize negative should be empty")
	}
	// input unchanged
	if !reflect.DeepEqual(in, []int{1, 2, 3, 4, 5}) {
		t.Fatal("SampleSize mutated input")
	}
}

func TestInvokeMap(t *testing.T) {
	got := InvokeMap([]string{"a-b", "c-d"}, func(s string, args ...any) []string {
		sep := args[0].(string)
		return strings.Split(s, sep)
	}, "-")
	want := [][]string{{"a", "b"}, {"c", "d"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("InvokeMap = %v", got)
	}
}

func TestNilRandDefaults(t *testing.T) {
	// Should not panic and should produce valid results with default rand.
	in := []int{1, 2, 3}
	if len(Shuffle(in, nil)) != 3 {
		t.Fatal("Shuffle nil rand failed")
	}
	if _, ok := Sample(in, nil); !ok {
		t.Fatal("Sample nil rand failed")
	}
	if len(SampleSize(in, 2, nil)) != 2 {
		t.Fatal("SampleSize nil rand failed")
	}
}
