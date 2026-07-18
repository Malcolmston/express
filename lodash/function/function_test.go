package function

import (
	"reflect"
	"strconv"
	"testing"
)

func TestAfter(t *testing.T) {
	calls := 0
	fn := After(3, func() int { calls++; return calls })
	if got := fn(); got != 0 {
		t.Errorf("call 1 = %d, want 0", got)
	}
	if got := fn(); got != 0 {
		t.Errorf("call 2 = %d, want 0", got)
	}
	if got := fn(); got != 1 {
		t.Errorf("call 3 = %d, want 1", got)
	}
	if got := fn(); got != 2 {
		t.Errorf("call 4 = %d, want 2", got)
	}
}

func TestBefore(t *testing.T) {
	calls := 0
	fn := Before(3, func() int { calls++; return calls })
	got := []int{fn(), fn(), fn(), fn()}
	// invoked on calls 1 and 2; calls 3,4 return last result (2)
	want := []int{1, 2, 2, 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Before = %v, want %v", got, want)
	}
	if calls != 2 {
		t.Errorf("underlying calls = %d, want 2", calls)
	}
}

func TestNegate(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }
	isOdd := Negate(isEven)
	if !isOdd(3) || isOdd(4) {
		t.Error("Negate")
	}
}

func TestMemoize(t *testing.T) {
	calls := 0
	sq := Memoize(func(n int) int { calls++; return n * n })
	if sq(4) != 16 || sq(4) != 16 || sq(5) != 25 {
		t.Error("Memoize wrong result")
	}
	if calls != 2 {
		t.Errorf("Memoize calls = %d, want 2", calls)
	}
}

func TestMemoizeWith(t *testing.T) {
	calls := 0
	type req struct{ id, ver int }
	fn := MemoizeWith(func(r req) int { calls++; return r.id }, func(r req) int { return r.id })
	fn(req{1, 1})
	fn(req{1, 2}) // same key -> cached
	fn(req{2, 1})
	if calls != 2 {
		t.Errorf("MemoizeWith calls = %d, want 2", calls)
	}
}

func TestFlip(t *testing.T) {
	sub := func(a, b int) int { return a - b }
	flipped := Flip(sub)
	if got := flipped(3, 10); got != 7 {
		t.Errorf("Flip = %d, want 7", got)
	}
}

func TestWrap(t *testing.T) {
	fn := Wrap("go", func(s string) string { return "<" + s + ">" })
	if got := fn(); got != "<go>" {
		t.Errorf("Wrap = %q", got)
	}
}

func TestOver(t *testing.T) {
	fn := Over(
		func(n int) int { return n + 1 },
		func(n int) int { return n * 2 },
	)
	if got := fn(5); !reflect.DeepEqual(got, []int{6, 10}) {
		t.Errorf("Over = %v", got)
	}
}

func TestOverEverySome(t *testing.T) {
	gt0 := func(n int) bool { return n > 0 }
	even := func(n int) bool { return n%2 == 0 }
	all := OverEvery(gt0, even)
	any := OverSome(gt0, even)
	if !all(4) || all(3) || all(-2) {
		t.Error("OverEvery")
	}
	if !any(3) || !any(-2) || any(-3) {
		t.Error("OverSome")
	}
}

func TestFlow(t *testing.T) {
	inc := func(n int) int { return n + 1 }
	dbl := func(n int) int { return n * 2 }
	if got := Flow(inc, dbl)(5); got != 12 { // (5+1)*2
		t.Errorf("Flow = %d, want 12", got)
	}
	if got := FlowRight(inc, dbl)(5); got != 11 { // (5*2)+1
		t.Errorf("FlowRight = %d, want 11", got)
	}
}

func TestCurry(t *testing.T) {
	add := func(a, b int) int { return a + b }
	if got := Curry2(add)(2)(3); got != 5 {
		t.Errorf("Curry2 = %d", got)
	}
	add3 := func(a, b, c int) int { return a + b + c }
	if got := Curry3(add3)(1)(2)(3); got != 6 {
		t.Errorf("Curry3 = %d", got)
	}
}

func TestPartial(t *testing.T) {
	concat := func(a, b string) string { return a + b }
	hi := Partial1(concat, "hi-")
	if got := hi("there"); got != "hi-there" {
		t.Errorf("Partial1 = %q", got)
	}
	f := func(a, b, c int) string { return strconv.Itoa(a + b + c) }
	g := Partial2(f, 1, 2)
	if got := g(3); got != "6" {
		t.Errorf("Partial2 = %q", got)
	}
}

func BenchmarkMemoize(b *testing.B) {
	sq := Memoize(func(n int) int { return n * n })
	for i := 0; i < b.N; i++ {
		_ = sq(i % 16)
	}
}
