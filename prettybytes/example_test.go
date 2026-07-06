package prettybytes_test

import (
	"fmt"

	"github.com/malcolmston/express/prettybytes"
)

// ExamplePrettyBytes renders a raw byte count as a compact, human-readable string
// using SI (base-1000) units, the default behaviour of the port. The value 1337
// is just over one kilobyte, so it is scaled into the "kB" unit and its mantissa
// is formatted with up to three significant digits, matching JavaScript's
// toPrecision(3). Trailing zeros are stripped and the integer part would be
// comma-grouped for larger magnitudes. This is the function to reach for when
// displaying file sizes or transfer amounts in a UI or log.
func ExamplePrettyBytes() {
	fmt.Println(prettybytes.PrettyBytes(1337))
	// Output: 1.34 kB
}

// ExamplePrettyBytes_zero shows the base-unit and boundary behaviour. Zero always
// renders as "0 B" in the plain (unsigned) form. The second call demonstrates
// that SI scaling divides by 1000, not 1024, so 1024 bytes is 1.02 kB rather than
// a round number. Both results follow pretty-bytes exactly, including the space
// between the number and its unit. These edge cases are worth knowing when the
// output must line up with a Node service using the original package.
func ExamplePrettyBytes_zero() {
	fmt.Println(prettybytes.PrettyBytes(0))
	fmt.Println(prettybytes.PrettyBytes(1024))
	// Output:
	// 0 B
	// 1.02 kB
}

// ExamplePrettyBytesOpts uses the Options struct to select alternative
// renderings. Setting Binary switches to base-1024 IEC units, so 1024 bytes
// becomes exactly "1 KiB". Setting Bits switches from byte units to bit units, so
// 1337 bytes is reported as "1.34 kbit". Each option composes independently, and
// combining Bits with Binary would yield the "kibit" family. This is the entry
// point whenever the default SI byte rendering is not the desired convention.
func ExamplePrettyBytesOpts() {
	fmt.Println(prettybytes.PrettyBytesOpts(1024, prettybytes.Options{Binary: true}))
	fmt.Println(prettybytes.PrettyBytesOpts(1337, prettybytes.Options{Bits: true}))
	// Output:
	// 1 KiB
	// 1.34 kbit
}
