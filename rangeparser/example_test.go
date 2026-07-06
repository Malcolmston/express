package rangeparser_test

import (
	"fmt"

	"github.com/malcolmston/express/rangeparser"
)

// ExampleParse resolves a simple byte range against a known resource size. The
// header "bytes=0-499" asks for the first 500 bytes, which Parse returns as a
// single inclusive Range from offset 0 to offset 499. The result code is
// ResultOK (0), signaling that at least one satisfiable range was produced.
// Both Start and End are inclusive, matching the HTTP specification. Callers
// branch on the returned code rather than on the slice length.
func ExampleParse() {
	ranges, code := rangeparser.Parse(1000, "bytes=0-499", false)
	fmt.Println(ranges, code)
	// Output: [{0 499}] 0
}

// ExampleParse_suffix shows a suffix range, where "-500" means the final 500
// bytes of the resource rather than an absolute offset. Against a 1000-byte
// resource this resolves to the inclusive range from offset 500 through 999.
// Parse translates the suffix form into absolute [Start, End] offsets so callers
// never have to do the arithmetic. The result code ResultOK (0) confirms the
// range is satisfiable. Suffix ranges are what let a client request "the last N
// bytes" without knowing the exact size.
func ExampleParse_suffix() {
	ranges, code := rangeparser.Parse(1000, "bytes=-500", false)
	fmt.Println(ranges, code)
	// Output: [{500 999}] 0
}

// ExampleParse_combine demonstrates the combine option, which merges
// overlapping and adjacent ranges into the minimal covering set. The first two
// specs overlap (0-100 and 50-200) and collapse into a single 0-200 range,
// while 300-400 stays separate because it does not touch the others. Merging
// preserves the order in which each merged group first appeared. This avoids
// emitting redundant or touching ranges to the caller. The result code is
// ResultOK (0).
func ExampleParse_combine() {
	ranges, code := rangeparser.Parse(1000, "bytes=0-100,50-200,300-400", true)
	fmt.Println(ranges, code)
	// Output: [{0 200} {300 400}] 0
}
