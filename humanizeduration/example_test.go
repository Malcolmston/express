package humanizeduration_test

import (
	"fmt"

	"github.com/malcolmston/express/humanizeduration"
)

// ExampleHumanize shows the default rendering of a millisecond count into an
// English phrase. Humanize takes an int64 number of milliseconds and returns the
// significant units joined by ", ", omitting any unit whose count is zero. Here
// 3,661,000 ms is exactly one hour, one minute and one second, so all three
// units appear. Counts of exactly one use the singular unit name while other
// counts are pluralized. The default unit set is years down to seconds, so
// sub-second remainders are dropped unless milliseconds are requested explicitly.
func ExampleHumanize() {
	fmt.Println(humanizeduration.Humanize(3661000))
	// Output: 1 hour, 1 minute, 1 second
}

// ExampleHumanize_zero demonstrates the empty-duration edge case. When the input
// reduces to nothing, Humanize does not return an empty string; instead it emits
// a single zero-valued phrase built from the smallest unit in effect. With the
// default unit set that smallest unit is seconds, so zero renders as "0 seconds".
// This mirrors the behaviour of the npm humanize-duration package. It keeps the
// output human-readable even for a zero duration.
func ExampleHumanize_zero() {
	fmt.Println(humanizeduration.Humanize(0))
	// Output: 0 seconds
}

// ExampleHumanizeOpts illustrates configuring the output through Options. The
// Largest field caps how many non-zero units are shown, so a duration of one
// hour, one minute and one second limited to two units keeps only the two
// largest. The Delimiter field replaces the default ", " separator, here with
// " and ". Fields left at their zero value fall back to the defaults, so Round,
// Spacer and Units are unchanged. This makes Options usable a la carte, setting
// only the knobs you care about.
func ExampleHumanizeOpts() {
	out := humanizeduration.HumanizeOpts(3661000, humanizeduration.Options{
		Largest:   2,
		Delimiter: " and ",
	})
	fmt.Println(out)
	// Output: 1 hour and 1 minute
}

// ExampleHumanizeOpts_milliseconds shows selecting an explicit unit set. By
// default milliseconds are never printed, but passing Units of {"s", "ms"}
// restricts the breakdown to seconds and milliseconds. The final unit in the set
// keeps any fractional remainder, though here 1500 ms divides cleanly into one
// second and 500 milliseconds. Every unit except the last is floored, so larger
// units always hold whole counts. This is how you surface millisecond precision
// that the default configuration hides.
func ExampleHumanizeOpts_milliseconds() {
	out := humanizeduration.HumanizeOpts(1500, humanizeduration.Options{
		Units: []string{"s", "ms"},
	})
	fmt.Println(out)
	// Output: 1 second, 500 milliseconds
}
