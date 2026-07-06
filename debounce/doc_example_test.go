package debounce_test

import (
	"fmt"
	"time"

	"github.com/malcolmston/express/debounce"
)

// ExampleDebouncer coalesces a burst of rapid calls into a single invocation.
// New wraps a function with a wait window; each Call within that window pushes
// the deadline back, so the wrapped function runs at most once per quiet period.
// A long wait is used here so the trailing timer never fires on its own, and
// Flush then invokes the single pending call synchronously. The counter is
// therefore incremented exactly once despite three calls.
func ExampleDebouncer() {
	count := 0
	d := debounce.New(time.Hour, func() { count++ })

	d.Call()
	d.Call()
	d.Call()
	d.Flush()

	fmt.Println(count)
	// Output: 1
}

// ExampleDebouncer_leading demonstrates leading-edge invocation. With
// WithLeading(true) and WithTrailing(false) the wrapped function fires
// immediately on the first Call of a burst and not again at the end. This is
// useful when the first event should be handled at once while a flood of
// following events is ignored until the window clears. Here only the first of
// three calls runs the function.
func ExampleDebouncer_leading() {
	count := 0
	d := debounce.New(time.Hour, func() { count++ },
		debounce.WithLeading(true),
		debounce.WithTrailing(false),
	)

	d.Call()
	d.Call()
	d.Call()

	fmt.Println(count)
	// Output: 1
}
