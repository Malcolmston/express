package dotenv_test

import (
	"fmt"

	"github.com/malcolmston/express/dotenv"
)

// ExampleBytes parses .env formatted content from a byte slice into a map of
// key/value pairs without touching the process environment. Blank lines and
// full-line comments starting with '#' are skipped, an optional "export " prefix
// is stripped so shell-sourceable files work, and each remaining line is split on
// the first '='. Double-quoted values have their backslash escapes expanded, so
// the "\n" in the greeting becomes a real newline (shown here with %q).
func ExampleBytes() {
	content := []byte("# app config\n" +
		"export HOST=localhost\n" +
		"PORT=8080\n" +
		"GREETING=\"hi\\nthere\"\n")

	env, err := dotenv.Bytes(content)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%q\n", env["HOST"])
	fmt.Printf("%q\n", env["PORT"])
	fmt.Printf("%q\n", env["GREETING"])
	// Output:
	// "localhost"
	// "8080"
	// "hi\nthere"
}

// ExampleParse_singleQuoted shows how quoting affects value handling. A
// single-quoted value is taken literally with no escape or variable expansion,
// so a backslash-n stays as the two characters "\n" rather than becoming a
// newline. An unquoted value has any trailing "#" comment removed and is then
// whitespace-trimmed. This example parses one single-quoted assignment from an
// in-memory reader.
func ExampleParse_singleQuoted() {
	env, err := dotenv.Bytes([]byte(`PATTERN='a\nb'`))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%q\n", env["PATTERN"])
	// Output: "a\\nb"
}
