package contenttype_test

import (
	"fmt"

	"github.com/malcolmston/express/contenttype"
)

// ExampleParse decodes a Content-Type header value into its media type and
// parameters. The value is split at the first ';' into the media type and a
// parameter list. The media type and parameter names are lower-cased, while
// parameter values keep their original case, and quoted values are unquoted.
// Here a typical header yields the media type and its charset parameter.
func ExampleParse() {
	ct, err := contenttype.Parse("text/html; charset=utf-8")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(ct.Type, ct.Parameters["charset"])
	// Output: text/html utf-8
}

// ExampleFormat serializes a ContentType back into a header value, the inverse
// of Parse. Parameters are emitted in sorted name order so the output is
// deterministic and stable to compare or snapshot. A parameter value that is a
// valid token is written unquoted; anything else is quoted and escaped. Here a
// JSON media type with a charset parameter is rendered into a header string.
func ExampleFormat() {
	s, err := contenttype.Format(contenttype.ContentType{
		Type:       "application/json",
		Parameters: map[string]string{"charset": "utf-8"},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(s)
	// Output: application/json; charset=utf-8
}
