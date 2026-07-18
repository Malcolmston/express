package str

import "strings"

// This file extends the str package with the few lodash "String" helpers the
// base file does not already cover: ToLower and ToUpper (plain whole-string case
// conversion, distinct from the word-splitting LowerCase/UpperCase), a Split
// with an optional result limit, and Chars, which splits a string into its
// Unicode characters. All are deterministic and depend only on the strings
// package plus Go's built-in rune handling.

// ToLower returns s with all Unicode letters mapped to lower case. Unlike
// LowerCase, it does not split the string into words; it mirrors lodash.toLower.
func ToLower(s string) string { return strings.ToLower(s) }

// ToUpper returns s with all Unicode letters mapped to upper case. Unlike
// UpperCase, it does not split the string into words; it mirrors lodash.toUpper.
func ToUpper(s string) string { return strings.ToUpper(s) }

// Split splits s around each instance of separator. When limit is negative all
// substrings are returned; when limit is zero the result is empty; otherwise at
// most limit substrings are returned, with the final one holding the unsplit
// remainder — matching lodash.split's limit semantics for whole-separator
// splitting.
func Split(s, separator string, limit int) []string {
	if limit == 0 {
		return []string{}
	}
	if limit < 0 {
		return strings.Split(s, separator)
	}
	return strings.SplitN(s, separator, limit)
}

// Chars splits s into a slice of its Unicode characters (runes as strings), so
// Chars("a€b") is ["a", "€", "b"]. It mirrors lodash's toArray/split("") on a
// string.
func Chars(s string) []string {
	runes := []rune(s)
	out := make([]string, len(runes))
	for i, r := range runes {
		out[i] = string(r)
	}
	return out
}
