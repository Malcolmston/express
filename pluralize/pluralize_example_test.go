package pluralize_test

import (
	"fmt"

	"github.com/malcolmston/express/pluralize"
)

// ExamplePlural converts singular words to their plural forms. It exercises the
// three tiers of the algorithm at once: regular suffix rules turn "test" into
// "tests" and "lady" into "ladies", explicit irregular pairs turn "person" into
// "people" and "mouse" into "mice", and uncountable words like "fish" are returned
// unchanged. Casing of the input is preserved, so a capitalized word stays
// capitalized. Each line prints the plural of the word shown.
func ExamplePlural() {
	fmt.Println(pluralize.Plural("test"))
	fmt.Println(pluralize.Plural("lady"))
	fmt.Println(pluralize.Plural("person"))
	fmt.Println(pluralize.Plural("mouse"))
	fmt.Println(pluralize.Plural("fish"))
	// Output:
	// tests
	// ladies
	// people
	// mice
	// fish
}

// ExampleSingular converts plural words back to their singular forms. It reverses
// the same rule tiers used by Plural: regular endings turn "boxes" into "box" and
// "cities" into "city", irregular plurals turn "children" into "child" and "teeth"
// into "tooth", and uncountable words such as "series" pass through untouched.
// Because the operation is idempotent, applying Singular to an already-singular
// word would leave it unchanged. Each line prints the singular of the word shown.
func ExampleSingular() {
	fmt.Println(pluralize.Singular("boxes"))
	fmt.Println(pluralize.Singular("cities"))
	fmt.Println(pluralize.Singular("children"))
	fmt.Println(pluralize.Singular("teeth"))
	fmt.Println(pluralize.Singular("series"))
	// Output:
	// box
	// city
	// child
	// tooth
	// series
}

// ExamplePlural_casePreservation highlights how the input's capitalization carries
// through to the output. An all-lower-case word yields a lower-case result, an
// all-upper-case word yields an upper-case result, and a word with a leading
// capital yields a capitalized result. This holds for irregular conversions too,
// so "Person" becomes "People" rather than "people". The logic mirrors the
// reference pluralize library's case-restoration behaviour. Each line prints one
// such conversion.
func ExamplePlural_casePreservation() {
	fmt.Println(pluralize.Plural("bus"))
	fmt.Println(pluralize.Plural("Bus"))
	fmt.Println(pluralize.Plural("BUS"))
	fmt.Println(pluralize.Plural("Person"))
	// Output:
	// buses
	// Buses
	// BUSES
	// People
}

// ExampleIsPlural and the companion IsSingular report which grammatical number a
// word is already in, without converting it. A countable word is in exactly one of
// the two forms, so IsPlural and IsSingular are complements for it. Uncountable
// words such as "fish" are considered to be both singular and plural because the
// two forms are identical, so both predicates return true. Here we probe a plural,
// a singular and an uncountable word. Each line prints a boolean.
func ExampleIsPlural() {
	fmt.Println(pluralize.IsPlural("people"))
	fmt.Println(pluralize.IsPlural("person"))
	fmt.Println(pluralize.IsSingular("person"))
	fmt.Println(pluralize.IsPlural("fish"))
	// Output:
	// true
	// false
	// true
	// true
}
