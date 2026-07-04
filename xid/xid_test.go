package xid

import (
	"testing"
)

func TestLength(t *testing.T) {
	s := New(1469918176)
	if len(s) != 20 {
		t.Fatalf("length = %d, want 20: %q", len(s), s)
	}
}

func TestTimeRoundTrip(t *testing.T) {
	machine := [3]byte{0x01, 0x02, 0x03}
	for _, sec := range []int64{0, 1, 1469918176, 2000000000} {
		s := NewWithData(sec, machine, 0x0405, 0x060708)
		got, err := Time(s)
		if err != nil {
			t.Fatalf("Time error: %v", err)
		}
		if got != sec {
			t.Fatalf("Time = %d, want %d", got, sec)
		}
	}
}

func TestDeterministic(t *testing.T) {
	machine := [3]byte{0x0a, 0x0b, 0x0c}
	a := NewWithData(1469918176, machine, 1234, 5678)
	b := NewWithData(1469918176, machine, 1234, 5678)
	if a != b {
		t.Fatalf("not deterministic: %q vs %q", a, b)
	}
}

func TestSortability(t *testing.T) {
	machine := [3]byte{0, 0, 0}
	small := NewWithData(1000, machine, 0, 0)
	large := NewWithData(2000, machine, 0, 0)
	if !(small < large) {
		t.Fatalf("expected %q < %q", small, large)
	}
}

func TestDecodeReversible(t *testing.T) {
	machine := [3]byte{0x11, 0x22, 0x33}
	s := NewWithData(1469918176, machine, 0x4455, 0x667788)
	b, err := Decode(s)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if encode(b) != s {
		t.Fatalf("round trip = %q, want %q", encode(b), s)
	}
	if b[4] != 0x11 || b[5] != 0x22 || b[6] != 0x33 {
		t.Fatalf("machine bytes wrong: %v", b[4:7])
	}
	if b[7] != 0x44 || b[8] != 0x55 {
		t.Fatalf("pid bytes wrong: %v", b[7:9])
	}
	if b[9] != 0x66 || b[10] != 0x77 || b[11] != 0x88 {
		t.Fatalf("counter bytes wrong: %v", b[9:12])
	}
}

func TestNewIncrements(t *testing.T) {
	a := New(1469918176)
	b := New(1469918176)
	if a == b {
		t.Fatalf("expected different ids from counter increment: %q", a)
	}
}
