package deburr_test

import (
	"fmt"

	"github.com/malcolmston/express/deburr"
)

// ExampleDeburr converts accented Latin letters to their basic ASCII
// equivalents and strips combining diacritical marks. Single letters map to a
// single letter, so "déjà" becomes "deja" and "Crème Brûlée" becomes "Creme
// Brulee". Characters that are neither accented Latin letters nor combining
// marks, such as the em dash here, pass through unchanged. This normalization is
// useful ahead of slugging, searching, or sorting.
func ExampleDeburr() {
	fmt.Println(deburr.Deburr("déjà vu — Crème Brûlée"))
	// Output: deja vu — Creme Brulee
}

// ExampleDeburr_ligatures shows that some letters expand to a short digraph
// rather than a single character, preserving the case of the original. The
// German sharp s becomes "ss", the ash ligature becomes "ae", and the thorn
// becomes "th". Because these expansions can lengthen the string, the output
// length is not tied to the input length. Here a word with a sharp s is
// deburred into its ASCII approximation.
func ExampleDeburr_ligatures() {
	fmt.Println(deburr.Deburr("straße"))
	// Output: strasse
}
