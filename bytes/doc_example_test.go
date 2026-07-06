package bytes_test

import (
	"fmt"

	"github.com/malcolmston/express/bytes"
)

// ExampleFormat renders a raw byte count into a compact human-readable label
// using the largest binary unit that fits. The unit ladder is binary, so 1024
// bytes is one kilobyte. By default up to two decimal places are shown with
// insignificant trailing zeros trimmed, so a whole kilobyte prints as "1KB"
// rather than "1.00KB". This is the form Express middleware uses when reporting
// sizes in logs and responses.
func ExampleFormat() {
	fmt.Println(bytes.Format(1024))
	fmt.Println(bytes.Format(1610612736))
	fmt.Println(bytes.Format(500))
	// Output:
	// 1KB
	// 1.5GB
	// 500B
}

// ExampleParse converts a human-written size string into an int64 count of
// bytes. It accepts an optional sign, an integer or decimal number, optional
// spaces, and a case-insensitive unit suffix; a bare number is interpreted as
// bytes. The parsed value is multiplied by the unit magnitude and floored, so
// "1.5MB" yields exactly 1572864. This is how a request body-size limit written
// as text becomes a number to compare against.
func ExampleParse() {
	n, err := bytes.Parse("1.5MB")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(n)
	// Output: 1572864
}

// ExampleFormatOpts shows the formatting knobs that mirror the npm options
// object. UnitSeparator inserts a string between the number and the unit, here a
// single space so the result reads "1.5 GB". Other options can fix the number of
// decimal places, keep trailing zeros, or force a specific unit. Passing the
// zero FormatOptions is equivalent to calling Format.
func ExampleFormatOpts() {
	fmt.Println(bytes.FormatOpts(1610612736, bytes.FormatOptions{UnitSeparator: " "}))
	// Output: 1.5 GB
}
