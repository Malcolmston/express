package ms_test

import (
	"fmt"
	"time"

	"github.com/malcolmston/express/ms"
)

// ExampleParse converts a human-readable duration string into a
// time.Duration. The input may use a short unit like "h" or a long unit like
// "days", optionally with a decimal number and surrounding spaces. Parsing is
// case-insensitive and understands common abbreviations. Here "2 days" becomes
// a 48-hour duration, which prints in Go's standard duration form. The second
// return value is a non-nil error only when the input cannot be parsed.
func ExampleParse() {
	d, err := ms.Parse("2 days")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(d)
	// Output: 48h0m0s
}

// ExampleParse_bareNumber shows the special rule that a number with no unit is
// interpreted as milliseconds, matching the npm ms package. This makes "100"
// mean 100 milliseconds rather than 100 of some larger unit. The rule is handy
// for configuration values that are already expressed in milliseconds. The
// parsed duration therefore prints as "100ms". Any explicit unit would
// override this default.
func ExampleParse_bareNumber() {
	d, _ := ms.Parse("100")
	fmt.Println(d)
	// Output: 100ms
}

// ExampleParse_negative demonstrates that Parse accepts a leading minus sign
// and decimal values together. The string "-1.5h" is one and a half hours in
// the past, which is ninety minutes. This is useful for expressing offsets
// relative to the present. The resulting duration prints in Go's canonical
// form as a negative hours-and-minutes value. Negative durations round-trip
// through the formatters as well.
func ExampleParse_negative() {
	d, _ := ms.Parse("-1.5h")
	fmt.Println(d)
	// Output: -1h30m0s
}

// ExampleFormat converts a time.Duration into the terse human-readable form,
// the direction the npm package uses without its { long } option. Format picks
// the largest unit whose absolute magnitude is at least one and rounds to the
// nearest whole count. A two-hour duration becomes "2h" and a sub-second
// duration stays in milliseconds. Negative durations keep their sign. The
// example prints several durations to show the unit selection.
func ExampleFormat() {
	fmt.Println(ms.Format(2 * time.Hour))
	fmt.Println(ms.Format(500 * time.Millisecond))
	fmt.Println(ms.Format(-3 * time.Second))
	// Output:
	// 2h
	// 500ms
	// -3s
}

// ExampleFormatLong converts a time.Duration into the spelled-out form, the
// direction the npm package selects with { long: true }. It chooses the same
// largest-fitting unit as Format but writes the unit name in full with correct
// pluralization. Matching the original, a value switches to the plural name
// once its magnitude reaches 1.5 units, so one hour stays singular while two
// hours is plural. The example prints a singular and a plural case together.
func ExampleFormatLong() {
	fmt.Println(ms.FormatLong(time.Hour))
	fmt.Println(ms.FormatLong(2 * time.Hour))
	fmt.Println(ms.FormatLong(time.Minute))
	// Output:
	// 1 hour
	// 2 hours
	// 1 minute
}
