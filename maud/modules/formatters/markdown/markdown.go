// basic Markdown converter
// ------------------------
// usage is as simple as htmlText := ParseMarkdown(markdownText)
// but note that this parser does not do anything for sanitizing HTML,
// so `markdownText` should have already been properly treated
// (e.g. disallowing dangerous tags like <script> and sanitizing HTML
// special characters).
// The parser is line-oriented, so it doesn't support multiline MD snippets.
// Currently, differently from standard Markdown, a newline gets converted in <br/>
// regardless of the trailing spaces.
// Special MD characters can be escaped via '\'.
package markdown

import (
	"log"
	"regexp"
	"strings"

	"github.com/hamcha/maud/maud/modules"
)

type mdPair struct {
	Regex *regexp.Regexp
	Func  func(*regexp.Regexp, string) string
}

var (
	mdElements []mdPair
	trimEscape *regexp.Regexp
)

func init() {
	// Order of regexes is important
	mdElements = []mdPair{
		{regexp.MustCompile(`(?U)(^|\\\\|[^\\])\*\*(.*[^\\])\*\*`), mdConvertTag("b")},
		{regexp.MustCompile(`(?U)(^|\\\\|[^\\\*])\*(.*[^\\])\*`), mdConvertTag("i")},
		{regexp.MustCompile(`(?U)(^|\\\\|[^\\])~~(.*[^\\])~~`), mdConvertTag("s")},
		{regexp.MustCompile(`(?U)(^|\\\\|[^\\])!\[(.*)\]\((.*)\)`), mdConvertImg},
		{regexp.MustCompile(`(?U)(^|\\\\|[^\\!])\[(.*)\]\((.*)\)`), mdConvertTagParam("a", "href")},
		{regexp.MustCompile("(?U)(^|\\\\\\\\|[^\\\\])`(.*[^\\\\])`"), mdConvertTag("code")},
	}
	trimEscape = regexp.MustCompile("\\\\([*~\\[!`\\\\])")
	log.Printf("[ OK ] Module initialized: Markdown")
}

type markdownFormatter struct{}

func Provide() modules.Formatter {
	return &markdownFormatter{}
}

// Format takes a markdown string as input and returns an HTML string.
// Allowed markdown snippets:
//   *italic*
//   **bold**
//   ~~strike~~
//   [alt](url)
//   ![alt](url)  -- insert image
//   `inline code`
func (m *markdownFormatter) Format(content string) string {
	lines := strings.Split(content, "\n")

	for idx := range lines {
		if len(strings.TrimSpace(lines[idx])) == 0 {
			continue
		}
		for _, pair := range mdElements {
			regex, fn := pair.Regex, pair.Func
			for regex.MatchString(lines[idx]) {
				lines[idx] = fn(regex, lines[idx])
			}
		}
		lines[idx] = trimEscape.ReplaceAllString(lines[idx], "$1")
	}

	return strings.Join(lines, "<br />\n")
}

// Unexported //

func mdConvertTag(tag string) func(*regexp.Regexp, string) string {
	return func(regex *regexp.Regexp, str string) string {
		return regex.ReplaceAllString(str, "$1<"+tag+">$2</"+tag+">")
	}
}

func mdConvertTagParam(tag, param string) func(*regexp.Regexp, string) string {
	return func(regex *regexp.Regexp, str string) string {
		return regex.ReplaceAllString(str, "$1<"+tag+" "+param+"=\"$3\">$2</"+tag+">")
	}
}

func mdConvertImg(regex *regexp.Regexp, str string) string {
	return regex.ReplaceAllString(str, `$1<a href="$3" rel="nofollow"><img src="$3" alt="$2"/></a>`)
}
