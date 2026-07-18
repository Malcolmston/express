package lang

import (
	"math"
	"testing"
)

func TestIsFiniteNaN(t *testing.T) {
	if !IsFinite(1.5) || IsFinite(math.Inf(1)) || IsFinite(math.NaN()) {
		t.Error("IsFinite")
	}
	if !IsNaN(math.NaN()) || IsNaN(1.0) {
		t.Error("IsNaN")
	}
}

func TestIsInteger(t *testing.T) {
	if !IsInteger(3.0) || IsInteger(3.5) || IsInteger(math.Inf(1)) {
		t.Error("IsInteger")
	}
}

func TestIsSafeInteger(t *testing.T) {
	if !IsSafeInteger(9007199254740991) {
		t.Error("IsSafeInteger max")
	}
	if IsSafeInteger(9007199254740992) {
		t.Error("IsSafeInteger over")
	}
	if IsSafeInteger(3.5) {
		t.Error("IsSafeInteger frac")
	}
}

func TestToSafeInteger(t *testing.T) {
	if ToSafeInteger(3.9) != 3 {
		t.Error("ToSafeInteger trunc")
	}
	if ToSafeInteger(math.NaN()) != 0 {
		t.Error("ToSafeInteger NaN")
	}
	if ToSafeInteger(1e20) != maxSafeInteger {
		t.Error("ToSafeInteger clamp")
	}
}

func TestToLength(t *testing.T) {
	if ToLength(3.7) != 3 {
		t.Error("ToLength trunc")
	}
	if ToLength(-5) != 0 {
		t.Error("ToLength negative")
	}
	if ToLength(1e20) != maxArrayLength {
		t.Error("ToLength clamp")
	}
}
