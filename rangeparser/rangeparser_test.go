package rangeparser

import (
	"reflect"
	"testing"
)

func TestSingleRange(t *testing.T) {
	got, res := Parse(1000, "bytes=0-499", false)
	if res != ResultOK {
		t.Fatalf("res = %d", res)
	}
	want := []Range{{Start: 0, End: 499}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSuffixRange(t *testing.T) {
	got, res := Parse(1000, "bytes=-500", false)
	if res != ResultOK {
		t.Fatalf("res = %d", res)
	}
	want := []Range{{Start: 500, End: 999}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestOpenEndedRange(t *testing.T) {
	got, res := Parse(1000, "bytes=500-", false)
	if res != ResultOK {
		t.Fatalf("res = %d", res)
	}
	want := []Range{{Start: 500, End: 999}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMultipleRanges(t *testing.T) {
	got, res := Parse(1000, "bytes=0-99,200-299", false)
	if res != ResultOK {
		t.Fatalf("res = %d", res)
	}
	want := []Range{{Start: 0, End: 99}, {Start: 200, End: 299}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestUnsatisfiable(t *testing.T) {
	_, res := Parse(200, "bytes=500-999", false)
	if res != ResultUnsatisfiable {
		t.Fatalf("res = %d want %d", res, ResultUnsatisfiable)
	}
}

func TestMalformedNoEquals(t *testing.T) {
	_, res := Parse(1000, "0-499", false)
	if res != ResultMalformed {
		t.Fatalf("res = %d want %d", res, ResultMalformed)
	}
}

func TestMalformedWrongUnitStillParsesType(t *testing.T) {
	// A non-bytes unit is still parsed structurally; expose the type.
	r, res := ParseRanges(1000, "items=0-5", false)
	if res != ResultOK {
		t.Fatalf("res = %d", res)
	}
	if r.Type != "items" {
		t.Fatalf("type = %q", r.Type)
	}
}

func TestMalformedBadNumbers(t *testing.T) {
	_, res := Parse(1000, "bytes=abc-def", false)
	if res != ResultMalformed {
		t.Fatalf("res = %d want %d", res, ResultMalformed)
	}
}

func TestCombineOverlapping(t *testing.T) {
	got, res := Parse(1000, "bytes=0-100,50-150,300-400", true)
	if res != ResultOK {
		t.Fatalf("res = %d", res)
	}
	want := []Range{{Start: 0, End: 150}, {Start: 300, End: 400}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestCombineAdjacent(t *testing.T) {
	got, res := Parse(1000, "bytes=0-99,100-199", true)
	if res != ResultOK {
		t.Fatalf("res = %d", res)
	}
	want := []Range{{Start: 0, End: 199}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestCombinePreservesOrder(t *testing.T) {
	// Out-of-order input; combined output ordered by first occurrence.
	got, res := Parse(1000, "bytes=300-400,0-100,50-150", true)
	if res != ResultOK {
		t.Fatalf("res = %d", res)
	}
	want := []Range{{Start: 300, End: 400}, {Start: 0, End: 150}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestType(t *testing.T) {
	r, res := ParseRanges(1000, "bytes=0-499", false)
	if res != ResultOK {
		t.Fatalf("res = %d", res)
	}
	if r.Type != "bytes" {
		t.Fatalf("type = %q want bytes", r.Type)
	}
}
