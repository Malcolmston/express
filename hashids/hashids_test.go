package hashids

import (
	"reflect"
	"testing"
)

func newHasher(t *testing.T) *HashID {
	t.Helper()
	h, err := New("this is my salt", 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return h
}

func TestOfficialVectors(t *testing.T) {
	h := newHasher(t)

	cases := []struct {
		nums []int64
		want string
	}{
		{[]int64{12345}, "NkK9"},
		{[]int64{1, 2, 3}, "laHquq"},
		{[]int64{683, 94108, 123, 5}, "aBMswoO2UB3Sj"},
	}

	for _, c := range cases {
		got, err := h.Encode(c.nums...)
		if err != nil {
			t.Fatalf("Encode(%v): %v", c.nums, err)
		}
		if got != c.want {
			t.Errorf("Encode(%v) = %q, want %q", c.nums, got, c.want)
		}
		back, err := h.Decode(got)
		if err != nil {
			t.Fatalf("Decode(%q): %v", got, err)
		}
		if !reflect.DeepEqual(back, c.nums) {
			t.Errorf("Decode(%q) = %v, want %v", got, back, c.nums)
		}
	}
}

func TestRoundTripSingle(t *testing.T) {
	h := newHasher(t)
	for _, n := range []int64{0, 1, 100, 999999, 1234567890} {
		enc, err := h.Encode(n)
		if err != nil {
			t.Fatal(err)
		}
		dec, err := h.Decode(enc)
		if err != nil {
			t.Fatal(err)
		}
		if len(dec) != 1 || dec[0] != n {
			t.Errorf("round trip %d: enc=%q dec=%v", n, enc, dec)
		}
	}
}

func TestMinLength(t *testing.T) {
	h, err := New("this is my salt", 18)
	if err != nil {
		t.Fatal(err)
	}
	enc, err := h.Encode(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(enc) < 18 {
		t.Errorf("min length not honored: %q (len %d)", enc, len(enc))
	}
	dec, err := h.Decode(enc)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dec, []int64{1}) {
		t.Errorf("Decode = %v, want [1]", dec)
	}
}

func TestEmptyEncode(t *testing.T) {
	h := newHasher(t)
	got, err := h.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("Encode() = %q, want empty", got)
	}
}

func TestDecodeInvalid(t *testing.T) {
	h := newHasher(t)
	got, err := h.Decode("this-is-not-valid!!!")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("Decode(invalid) = %v, want empty", got)
	}
}

func TestNegativeRejected(t *testing.T) {
	h := newHasher(t)
	if _, err := h.Encode(-1); err == nil {
		t.Fatal("expected error for negative number")
	}
}
