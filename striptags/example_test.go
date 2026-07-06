package striptags_test

import (
	"fmt"

	"github.com/malcolmston/express/striptags"
)

// ExampleStripTags removes every HTML tag while preserving the text between
// them. With no allowed tags supplied the paragraph and bold markup are both
// stripped, leaving only the plain text content. The algorithm is a single-pass
// state machine, not a full HTML parser, but it handles well-formed markup
// cleanly. This is the common case for generating a plain-text preview from a
// fragment of HTML. HTML comments would also be removed regardless of any
// allowed list.
func ExampleStripTags() {
	fmt.Println(striptags.StripTags("<p>Hello <b>World</b></p>"))
	// Output: Hello World
}

// ExampleStripTags_allowed keeps a whitelist of tags while removing all others.
// Passing "b" preserves both the opening and closing bold tags verbatim, so the
// emphasis survives, while the surrounding paragraph tags are stripped. Allowed
// tag names may be written bare ("b") or wrapped in angle brackets ("<b>"), and
// matching is case-insensitive. Attributes on an allowed tag are preserved
// exactly as written. This lets callers permit a small set of formatting tags in
// otherwise untrusted text.
func ExampleStripTags_allowed() {
	fmt.Println(striptags.StripTags("<p>Hi <b>there</b></p>", "b"))
	// Output: Hi <b>there</b>
}
