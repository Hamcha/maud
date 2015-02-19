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
	"../.."
	"regexp"
	"strings"
)

func Provide() modules.Formatter {
	md := new(MarkdownFormatter)
	md.Init()
	return md
}

type MarkdownFormatter struct {
	mdElements map[*regexp.Regexp]func(*regexp.Regexp, string) string
	trimEscape *regexp.Regexp
}

func (m *MarkdownFormatter) Init() {
	m.mdElements = map[*regexp.Regexp]func(*regexp.Regexp, string) string{
		regexp.MustCompile(`(?U)(^|\\\\|[^\\])\*\*(.*[^\\])\*\*`):   mdConvertTag("b"),
		regexp.MustCompile(`(?U)(^|\\\\|[^\\])\*(.*[^\\])\*`):       mdConvertTag("i"),
		regexp.MustCompile(`(?U)(^|\\\\|[^\\])!\[(.*)\]\((.*)\)`):   mdConvertImg,
		regexp.MustCompile(`(?U)(^|\\\\|[^\\!])\[(.*)\]\((.*)\)`):   mdConvertTagParam("a", "href"),
		regexp.MustCompile("(?U)(^|\\\\\\\\|[^\\\\])`(.*[^\\\\])`"): mdConvertTag("code"),
	}
	m.trimEscape = regexp.MustCompile("\\\\([*\\[!`\\\\])")
}

// Allowed markdown snippets:
//   *italic*
//   **bold**
//   [alt](url)
//   ![alt](url)  -- insert image
//   `inline code`
func (m *MarkdownFormatter) Format(content string) string {
	lines := strings.Split(content, "\n")

	for idx := range lines {
		if len(strings.TrimSpace(lines[idx])) == 0 {
			continue
		}
		for regex, fn := range m.mdElements {
			for regex.MatchString(lines[idx]) {
				lines[idx] = fn(regex, lines[idx])
			}
		}
		lines[idx] = m.trimEscape.ReplaceAllString(lines[idx], "$1")
	}

	return strings.Join(lines, "<br />\n")
}

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
	return regex.ReplaceAllString(str, "$1<img src=\"$3\" alt=\"$2\"/>")
}

func mdConvertQuote(_ *regexp.Regexp, str string) string {
	return "<span class=\"purpletext\">&gt; " + strings.TrimSpace(str[4:]) + "</span>"
}

func mdConvertPostQuote(regex *regexp.Regexp, str string) string {
	return regex.ReplaceAllString(str, "<a href=\"#p$1\" class=\"postIdQuote\">$0</a>")
}
