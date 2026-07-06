package mediatyper_test

import (
	"fmt"

	"github.com/malcolmston/express/mediatyper"
)

// ExampleParse shows how a full media type string is broken into its parts.
// The input here carries a top-level type, a subtype with a structured
// "+json" suffix, and a charset parameter. Parse lower-cases the type,
// subtype and suffix, exposes the suffix separately from the base subtype,
// and unescapes parameter values into a map keyed by lower-case name. The
// second return value is a non-nil error only when the input is malformed.
// Printing the individual fields shows exactly how the string was decomposed.
func ExampleParse() {
	m, err := mediatyper.Parse("application/vnd.api+json; charset=utf-8")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("type:", m.Type)
	fmt.Println("subtype:", m.Subtype)
	fmt.Println("suffix:", m.Suffix)
	fmt.Println("charset:", m.Parameters["charset"])
	// Output:
	// type: application
	// subtype: vnd.api
	// suffix: json
	// charset: utf-8
}

// ExampleFormat shows the inverse of Parse: turning a MediaType value back
// into a canonical string. The type, subtype and suffix are validated as
// tokens and lower-cased, and any parameters are emitted in a stable sorted
// order so the result is deterministic. A parameter value that is a plain
// token is written as-is, while a non-token value would be quoted. Format
// returns an error only when a component is not a legal token, which is not
// the case here.
func ExampleFormat() {
	s, err := mediatyper.Format(mediatyper.MediaType{
		Type:       "text",
		Subtype:    "html",
		Parameters: map[string]string{"charset": "utf-8"},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(s)
	// Output: text/html; charset=utf-8
}
