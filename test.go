package main

import (
	"fmt"
	"strings"
)

var bbElements map[string]func(string, string) string

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
			tag = tag[1:]
			if stack[top].Name == tag {
				content := code[stack[top].End:start]
				parsed := bbElements[tag](stack[top].Parameter, content)
				code = code[0:stack[top].Start] + parsed + code[offset:]
				// Pop stack
				stack = stack[:top]
				top -= 1
			}
		} else {
			// Separate parameter, if given
			parameter := ""
			if index(tag, 0, '=') > 0 {
				parts := strings.SplitN(tag, "=", 2)
				tag = parts[0]
				parameter = parts[1]
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

func main() {
	bbElements = make(map[string]func(string, string) string)
	bbElements["b"] = func(par, con string) string {
		return "<b>" + con + "</b>"
	}
	bbElements["img"] = func(par, con string) string {
		return "<img src=\"" + con + "\" />"
	}
	bbElements["url"] = func(par, con string) string {
		return "<a href=\"" + par + "\">" + con + "</a>"
	}
	in := "[b]Ciao mamma[/b] ho trovato [url=imgur.com]sta roba [img]lol.jpg[/img][/url]"
	fmt.Println("In: " + in)
	fmt.Println("Out: " + bbcode(in))
}
