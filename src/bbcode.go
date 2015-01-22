package main

import (
	"net/url"
	"strings"
)

var bbElements map[string]func(string, string) string

func initbbcode() {
	bbElements = make(map[string]func(string, string) string)
	// Standard BBcode -> HTML tags
	bbElements["b"] = bbToHTML("b")
	bbElements["i"] = bbToHTML("i")
	bbElements["u"] = bbToHTML("u")
	bbElements["strike"] = bbToHTML("s")
	// Other BBcode tags
	bbElements["img"] = func(_, con string) string {
		idx := strings.IndexRune(con, '?')
		if idx > 0 {
			con = con[0:idx] + url.QueryEscape(con[idx:])
		}
		return "<a href=\"" + con + "\"><img src=\"" + con + "\" /></a>"
	}
	bbElements["url"] = func(par, con string) string {
		if len(par) < 1 {
			par = con
		}
		// if content is already a hyperlink, just return it
		if strings.HasPrefix(par, "<a ") {
			return par
		}
		if !strings.HasPrefix(par, "http://") && !strings.HasPrefix(par, "https://") {
			par = "http://" + par
		}
		return "<a href=\"" + par + "\">" + con + "</a>"
	}
	bbElements["spoiler"] = func(_, con string) string {
		return "<span class=\"spoiler\">" + con + "</span>"
	}
	bbElements["youtube"] = func(_, con string) string {
		idx := strings.Index(con, "?v=")
		if idx > 0 {
			con = con[idx+3:]
		}
		return "<iframe width=\"560\" height=\"315\" src=\"//www.youtube.com/embed/" + url.QueryEscape(con) + "\" frameborder=\"0\" allowfullscreen></iframe>"
	}
}

func bbToHTML(tag string) func(string, string) string {
	return func(_, con string) string {
		return "<" + tag + ">" + con + "</" + tag + ">"
	}
}

func index(str string, offset int, del uint8) int {
	for i := offset; i < len(str); i++ {
		if str[i] == del {
			return i
		}
	}
	return -1
}

func bbcode(code string) string {
	offset := 0
	type BBCode struct {
		Name      string
		Parameter string
		Start     int
		End       int
	}
	stack := make([]BBCode, 0)
	top := -1
	for {
		// Get next tag in string (Regexp free)
		start := index(code, offset, '[')
		if start < 0 {
			break
		}
		end := index(code, start+1, ']')
		if end < 0 {
			break
		}
		offset = end + 1
		tag := code[start+1 : end]

		// Is it a closing tag?
		if top >= 0 && tag[0] == '/' {
			tag = strings.ToLower(tag[1:])
			for idx := top; idx >= 0; idx -= 1 {
				if stack[idx].Name == tag {
					content := code[stack[top].End:start]
					parsed := bbElements[tag](stack[top].Parameter, content)
					code = code[0:stack[top].Start] + parsed + code[offset:]
					// Pop stack
					stack = stack[:idx]
					top = idx - 1
					break
				}
			}
		} else {
			// Separate parameter, if given
			parameter := ""
			if index(tag, 0, '=') > 0 {
				parts := strings.SplitN(tag, "=", 2)
				tag = strings.ToLower(parts[0])
				parameter = parts[1]
			} else {
				tag = strings.ToLower(tag)
			}
			// Is it a registered bbcode?
			if _, ok := bbElements[tag]; ok {
				stack = append(stack, BBCode{
					Name:      tag,
					Parameter: parameter,
					Start:     start,
					End:       offset,
				})
				top += 1
			}
		}
	}

	return code
}
