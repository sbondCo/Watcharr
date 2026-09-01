package util

import (
	"bytes"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

// Convert the given Markdown string to HTML and sanitize it.
//
// If the conversion fails, it returns the plain Markdown input, assuming that the input just isn't in Markdown format or already HTML formatted.
func MdToHTMLSafe(markdown string) string {
	// try to convert to html
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &buf); err != nil {
		return markdown
	}

	// sanitize output, i.e. remove JavaScript or other potentially malicious code
	policy := bluemonday.UGCPolicy()
	return policy.Sanitize(buf.String())
}
