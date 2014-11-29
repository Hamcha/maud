// basic Markdown converter
// ------------------------
// usage is as simple as htmlText := ParseMarkdown(markdownText)
// but note that this parser does not do anything for sanitizing HTML,
// so `markdownText` should have already been properly treated
// (e.g. disallowing dangerous tags like <script> and sanitizing HTML
// special characters).
// HOWEVER: leading ">" should NOT be converted to &gt; or else the
// line won't be accounted as quote.
// The parser is line-oriented, so it doesn't support multiline MD snippets.
// Currently, differently from standard Markdown, a newline gets converted in <br/>
// regardless of the trailing spaces.
// Special MD characters can be escaped via '\'.
package main

import (
	"regexp"
	"strings"
)

var mdElements map[*regexp.Regexp]func(*regexp.Regexp, string) string
var trimEscape *regexp.Regexp

func initMarkdown() {
	mdElements = map[*regexp.Regexp]func(*regexp.Regexp, string) string{
		regexp.MustCompile("(?U)(^|\\\\\\\\|[^\\\\])\\*\\*(.*[^\\\\])\\*\\*"): mdConvertTag("b"),
		regexp.MustCompile("(?U)(^|\\\\\\\\|[^\\\\])\\*(.*[^\\\\])\\*"):       mdConvertTag("i"),
		regexp.MustCompile("(?U)(^|\\\\\\\\|[^\\\\])!\\[(.*)\\]\\((.*)\\)"):   mdConvertTagParam("iframe", "src"),
		regexp.MustCompile("(?U)(^|\\\\\\\\|[^\\\\])\\[(.*)\\]\\((.*)\\)"):    mdConvertTagParam("a", "href"),
		regexp.MustCompile("(?U)(^|\\\\\\\\|[^\\\\])`(.*[^\\\\])`"):           mdConvertTag("code"),
		regexp.MustCompile("^&gt;.*$"):                                        mdConvertQuote,
	}
	trimEscape = regexp.MustCompile("\\\\([*\\[!`\\\\])")
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

func mdConvertQuote(regex *regexp.Regexp, str string) string {
	return "<span class=\"purpletext\">&gt; " + strings.TrimSpace(str[4:]) + "</span>"
}

// Allowed markdown snippets:
//   *italic*
//   **bold**
//   [alt](url)
//   ![alt](url)  -- embeds resource
//   >quote
//   `inline code`
func ParseMarkdown(content string) string {
	lines := strings.Split(content, "\n")

	for idx := range lines {
		if len(strings.TrimSpace(lines[idx])) == 0 {
			continue
		}
		for regex, fn := range mdElements {
			for regex.MatchString(lines[idx]) {
				lines[idx] = fn(regex, lines[idx])
			}
		}
		lines[idx] = trimEscape.ReplaceAllString(lines[idx], "$1")
	}

	return strings.Join(lines, "<br />\n")
}
