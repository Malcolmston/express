package filesize_test

import (
	"fmt"

	"github.com/malcolmston/express/filesize"
)

// ExampleFileSize renders a raw byte count with the default options, which use
// base 10 (SI) units and two fractional digits. The value 1337 falls between
// one and two kilobytes, so it is divided by 1000 and shown in kB. The quotient
// 1.337 is rounded to two decimals, yielding "1.34 kB". This is the convenient
// entry point when the defaults are acceptable and no unit family override is
// needed. The unit label is separated from the number by a single space.
func ExampleFileSize() {
	fmt.Println(filesize.FileSize(1337))
	// Output: 1.34 kB
}

// ExampleFileSize_roundValues shows how trailing zeros are stripped. Exactly one
// thousand bytes divides evenly into one kilobyte, so instead of "1.00 kB" the
// result is the cleaner "1 kB". A value of 1500 divides to 1.5, and the dangling
// zero in "1.50" is removed to give "1.5 kB". Zero is a special case that always
// renders as "0 B". Negative inputs are formatted by magnitude with a leading
// minus sign.
func ExampleFileSize_roundValues() {
	fmt.Println(filesize.FileSize(1000))
	fmt.Println(filesize.FileSize(1500))
	fmt.Println(filesize.FileSize(0))
	fmt.Println(filesize.FileSize(-1337))
	// Output:
	// 1 kB
	// 1.5 kB
	// 0 B
	// -1.34 kB
}

// ExampleFileSizeOpts demonstrates overriding the defaults through Options.
// Setting Base to 2 switches the divisor to 1024 and, since Standard is left
// empty, derives the IEC unit family (KiB, MiB, ...). The same 1337 bytes is now
// just over one kibibyte, rendering as "1.31 KiB". The jedec standard also uses
// the 1024 divisor but spells the units in the SI style (KB, MB, ...), so 1536
// bytes becomes "1.5 KB". This is the entry point to use when a specific
// convention must be matched.
func ExampleFileSizeOpts() {
	fmt.Println(filesize.FileSizeOpts(1337, filesize.Options{Base: 2}))
	fmt.Println(filesize.FileSizeOpts(1536, filesize.Options{Base: 2, Standard: "jedec"}))
	// Output:
	// 1.31 KiB
	// 1.5 KB
}

// ExampleFileSizeOpts_round controls the number of fractional digits with the
// Round option, which points at an int so that zero is distinguishable from
// unset. A Round of 0 truncates 1337 bytes to a whole "1 kB". A Round of 3 keeps
// three decimals, exposing the full "1.337 kB". Trailing-zero stripping still
// applies after rounding, so a value that lands on a round number loses its
// decimals regardless of the requested precision. This makes Round an upper
// bound on the digits shown rather than a fixed width.
func ExampleFileSizeOpts_round() {
	zero := 0
	three := 3
	fmt.Println(filesize.FileSizeOpts(1337, filesize.Options{Round: &zero}))
	fmt.Println(filesize.FileSizeOpts(1337, filesize.Options{Round: &three}))
	// Output:
	// 1 kB
	// 1.337 kB
}
