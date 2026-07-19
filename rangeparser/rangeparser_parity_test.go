package rangeparser

import (
	"reflect"
	"testing"
)

// Parity tests for the jshttp/range-parser upstream library.
//
// All input -> expected-output vectors below are transcribed verbatim from the
// upstream test suite and reference implementation:
//
//	https://raw.githubusercontent.com/jshttp/range-parser/master/test/range-parser.js
//	https://raw.githubusercontent.com/jshttp/range-parser/master/index.js
//
// Upstream returns either the numeric sentinels -1 (unsatisfiable) / -2
// (malformed), mapped here to ResultUnsatisfiable / ResultMalformed, or an
// array of {start, end} objects carrying a `type`, mapped here to
// []Range with Ranges.Type.

// TestParityMalformed covers every upstream vector that returns -2.
func TestParityMalformed(t *testing.T) {
	cases := []struct {
		size   int64
		header string
	}{
		{200, ""},
		{200, "bytes=100200"},
		{200, "bytes=,100200"},
		{200, "malformed"},
		{200, "bytes=x-100"},
		{200, "bytes=100-x"},
		{200, "bytes=--100"},
		{200, "bytes=100--200"},
		{200, "bytes=-"},
		{200, "bytes= - "},
		{200, "bytes="},
		{200, "bytes=,"},
		{200, "bytes= , , "},
		{200, "bytes=100-200-300"},
		{200, "bytes=-100-150"},
		{200, "bytes=01a-150"},
		{200, "bytes=100-15b0"},
		{200, "bytes=y-v,x-"},
		{200, "bytes=abc-def,ghi-jkl"},
		{200, "bytes=x-,y-,z-"},
	}
	for _, c := range cases {
		_, res := Parse(c.size, c.header, false)
		if res != ResultMalformed {
			t.Errorf("Parse(%d, %q) = %d, want ResultMalformed (%d)", c.size, c.header, res, ResultMalformed)
		}
	}
}

// TestParityUnsatisfiable covers every upstream vector that returns -1.
func TestParityUnsatisfiable(t *testing.T) {
	cases := []struct {
		size   int64
		header string
	}{
		{200, "bytes=500-600"},
		{200, "bytes=500-600,601-700"},
		{200, "bytes=500-20"},
		{200, "bytes=500-999"},
		{200, "bytes=500-999,1000-1499"},
		{200, "bytes=abc-def,500-999"},
		{200, "bytes=500-999,xyz-uvw"},
		{200, "bytes=abc-def,500-999,xyz-uvw"},
		{1000, "bytes=-0"},
	}
	for _, c := range cases {
		_, res := Parse(c.size, c.header, false)
		if res != ResultUnsatisfiable {
			t.Errorf("Parse(%d, %q) = %d, want ResultUnsatisfiable (%d)", c.size, c.header, res, ResultUnsatisfiable)
		}
	}
}

// TestParseParityValid covers every upstream vector that returns ranges (no
// combine).
func TestParityValid(t *testing.T) {
	cases := []struct {
		size   int64
		header string
		want   []Range
	}{
		{1000, "bytes=0-499", []Range{{0, 499}}},
		{200, "bytes=0-499", []Range{{0, 199}}}, // cap end at size
		{1000, "bytes=40-80", []Range{{40, 80}}},
		{1000, "bytes=-400", []Range{{600, 999}}}, // last n bytes
		{100, "bytes=-101", []Range{{0, 99}}},     // suffix larger than size
		{1000, "bytes=400-", []Range{{400, 999}}}, // only start
		{1000, "bytes=0-", []Range{{0, 999}}},
		{1000, "bytes=0-0", []Range{{0, 0}}},
		{1000, "bytes=-1", []Range{{999, 999}}},         // last byte
		{1000, "bytes=100-200,x-", []Range{{100, 200}}}, // ignore invalid when valid exists
		{1000, "bytes=x-,0-100,y-", []Range{{0, 100}}},
		{1000, "bytes=0-50,abc-def,100-150", []Range{{0, 50}, {100, 150}}},
		{1000, "bytes=40-80,81-90,-1", []Range{{40, 80}, {81, 90}, {999, 999}}},
		{200, "bytes=0-499,1000-,500-999", []Range{{0, 199}}},
		{1000, "bytes=   40-80 , 81-90 , -1 ", []Range{{40, 80}, {81, 90}, {999, 999}}},
		{1000, "items=0-5", []Range{{0, 5}}},
		{1000, "bytes= , 0-10", []Range{{0, 10}}},
	}
	for _, c := range cases {
		got, res := Parse(c.size, c.header, false)
		if res != ResultOK {
			t.Errorf("Parse(%d, %q) res = %d, want ResultOK", c.size, c.header, res)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Parse(%d, %q) = %v, want %v", c.size, c.header, got, c.want)
		}
	}
}

// TestParityType checks the reported unit for representative valid vectors.
func TestParityType(t *testing.T) {
	cases := []struct {
		size   int64
		header string
		typ    string
	}{
		{1000, "bytes=0-499", "bytes"},
		{1000, "items=0-5", "items"},
	}
	for _, c := range cases {
		r, res := ParseRanges(c.size, c.header, false)
		if res != ResultOK {
			t.Errorf("ParseRanges(%d, %q) res = %d, want ResultOK", c.size, c.header, res)
			continue
		}
		if r.Type != c.typ {
			t.Errorf("ParseRanges(%d, %q) type = %q, want %q", c.size, c.header, r.Type, c.typ)
		}
	}
}

// TestParityCombine covers the combine:true vectors.
func TestParityCombine(t *testing.T) {
	cases := []struct {
		size   int64
		header string
		want   []Range
	}{
		{150, "bytes=0-4,90-99,5-75,100-199,101-102", []Range{{0, 75}, {90, 149}}},
		{150, "bytes=-1,20-100,0-1,101-120", []Range{{149, 149}, {20, 120}, {0, 1}}},
	}
	for _, c := range cases {
		got, res := Parse(c.size, c.header, true)
		if res != ResultOK {
			t.Errorf("Parse(%d, %q, combine) res = %d, want ResultOK", c.size, c.header, res)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Parse(%d, %q, combine) = %v, want %v", c.size, c.header, got, c.want)
		}
	}
}
