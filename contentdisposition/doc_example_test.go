package contentdisposition_test

import (
	"fmt"

	"github.com/malcolmston/express/contentdisposition"
)

// ExampleFormat builds a Content-Disposition header for a download with an
// ASCII filename. The default disposition type is "attachment", which tells a
// browser to save the body rather than display it. A pure-ASCII name is emitted
// as a single quoted filename parameter. This is what a server sends so the
// browser offers a sensible "Save As" name.
func ExampleFormat() {
	fmt.Println(contentdisposition.Format("report.pdf"))
	// Output: attachment; filename="report.pdf"
}

// ExampleFormat_unicode shows how a non-ASCII filename is encoded. Because a
// bare filename parameter may only carry ASCII, an RFC 5987 filename* parameter
// is added with the name UTF-8 percent-encoded. By default a legacy ASCII
// fallback is emitted alongside it, built by replacing each non-ASCII byte with
// '?'. Browsers that understand filename* use the decoded Unicode name and
// ignore the fallback.
func ExampleFormat_unicode() {
	fmt.Println(contentdisposition.Format("naïve.txt"))
	// Output: attachment; filename="na??ve.txt"; filename*=UTF-8''na%c3%afve.txt
}

// ExampleParse decodes a Content-Disposition header value back into its type and
// filename. The disposition type is lower-cased and quoted parameter values are
// unescaped. When both filename and filename* are present the decoded extended
// value wins, matching browser precedence. Here a simple attachment header is
// parsed into its type and suggested filename.
func ExampleParse() {
	cd, err := contentdisposition.Parse(`attachment; filename="report.pdf"`)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cd.Type, cd.Filename)
	// Output: attachment report.pdf
}
