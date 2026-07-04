package math

import (
	"math"
	"testing"
)

type obj struct {
	N int
}

func TestSumAndBy(t *testing.T) {
	if got := Sum([]int{4, 2, 8, 6}); got != 20 {
		t.Errorf("Sum = %d", got)
	}
	if got := Sum([]float64{1.5, 2.5}); got != 4.0 {
		t.Errorf("Sum float = %v", got)
	}
	objs := []obj{{6}, {4}, {2}}
	if got := SumBy(objs, func(o obj) int { return o.N }); got != 12 {
		t.Errorf("SumBy = %d", got)
	}
}

func TestMean(t *testing.T) {
	if got := Mean([]int{4, 2, 8, 6}); got != 5 {
		t.Errorf("Mean = %v", got)
	}
	objs := []obj{{6}, {4}, {2}}
	if got := MeanBy(objs, func(o obj) int { return o.N }); got != 4 {
		t.Errorf("MeanBy = %v", got)
	}
	if got := Mean([]int{}); !math.IsNaN(got) {
		t.Errorf("Mean empty = %v, want NaN", got)
	}
}

func TestMaxMin(t *testing.T) {
	if got := Max([]int{4, 2, 8, 6}); got != 8 {
		t.Errorf("Max = %d", got)
	}
	if got := Min([]int{4, 2, 8, 6}); got != 2 {
		t.Errorf("Min = %d", got)
	}
	if got := Max([]int{}); got != 0 {
		t.Errorf("Max empty = %d", got)
	}
	objs := []obj{{6}, {4}, {8}, {2}}
	if got := MaxBy(objs, func(o obj) int { return o.N }); got.N != 8 {
		t.Errorf("MaxBy = %+v", got)
	}
	if got := MinBy(objs, func(o obj) int { return o.N }); got.N != 2 {
		t.Errorf("MinBy = %+v", got)
	}
}

func TestClamp(t *testing.T) {
	if got := Clamp(-10, -5, 5); got != -5 {
		t.Errorf("Clamp low = %d", got)
	}
	if got := Clamp(10, -5, 5); got != 5 {
		t.Errorf("Clamp high = %d", got)
	}
	if got := Clamp(3, -5, 5); got != 3 {
		t.Errorf("Clamp mid = %d", got)
	}
	if got := Clamp(10, 5, -5); got != 5 {
		t.Errorf("Clamp swapped = %d", got)
	}
}

func TestInRange(t *testing.T) {
	cases := []struct {
		n, start, end int
		want          bool
	}{
		{3, 2, 4, true},
		{4, 8, 2, true},
		{2, 2, 4, true},
		{4, 2, 4, false},
		{-3, -2, -6, true},
	}
	for _, c := range cases {
		if got := InRange(c.n, c.start, c.end); got != c.want {
			t.Errorf("InRange(%d,%d,%d) = %v, want %v", c.n, c.start, c.end, got, c.want)
		}
	}
}

func TestRoundCeilFloor(t *testing.T) {
	if got := Round(4.006, 2); got != 4.01 {
		t.Errorf("Round = %v", got)
	}
	if got := Round(4.006, 0); got != 4 {
		t.Errorf("Round p0 = %v", got)
	}
	if got := Round(4060, -2); got != 4100 {
		t.Errorf("Round neg = %v", got)
	}
	if got := Ceil(6.004, 2); got != 6.01 {
		t.Errorf("Ceil = %v", got)
	}
	if got := Ceil(6040, -2); got != 6100 {
		t.Errorf("Ceil neg = %v", got)
	}
	if got := Floor(0.046, 2); got != 0.04 {
		t.Errorf("Floor = %v", got)
	}
	if got := Floor(4060, -2); got != 4000 {
		t.Errorf("Floor neg = %v", got)
	}
	// half rounds toward +Inf, matching JavaScript Math.round.
	if got := Round(2.5, 0); got != 3 {
		t.Errorf("Round 2.5 = %v", got)
	}
	if got := Round(-2.5, 0); got != -2 {
		t.Errorf("Round -2.5 = %v", got)
	}
}

func TestArithmetic(t *testing.T) {
	if got := Add(6, 4); got != 10 {
		t.Errorf("Add = %d", got)
	}
	if got := Subtract(6, 4); got != 2 {
		t.Errorf("Subtract = %d", got)
	}
	if got := Multiply(6, 4); got != 24 {
		t.Errorf("Multiply = %d", got)
	}
	if got := Divide(6.0, 4.0); got != 1.5 {
		t.Errorf("Divide = %v", got)
	}
}

func TestRandom(t *testing.T) {
	half := func() float64 { return 0.5 }
	if got := Random(0, 5, half); got != 2.5 {
		t.Errorf("Random = %v", got)
	}
	if got := Random(5, 0, half); got != 2.5 {
		t.Errorf("Random swapped = %v", got)
	}
	if got := RandomInt(0, 5, half); got != 3 {
		t.Errorf("RandomInt = %d", got)
	}
	// Lower bound reachable.
	if got := RandomInt(1, 5, func() float64 { return 0 }); got != 1 {
		t.Errorf("RandomInt low = %d", got)
	}
	// Upper bound reachable even when rnd approaches 1.
	if got := RandomInt(1, 5, func() float64 { return 0.9999999 }); got != 5 {
		t.Errorf("RandomInt high = %d", got)
	}
}
