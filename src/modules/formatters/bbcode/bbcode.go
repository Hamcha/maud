package bbcode

import (
	"../.."
	"net/url"
	"strings"
)

func Provide() modules.Formatter {
	bbcode := new(BBCodeFormatter)
	bbcode.Init()
	return bbcode
}

type BBCodeFormatter struct {
	bbElements map[string]func(string, string) string
}

func (b *BBCodeFormatter) Init() {
	b.bbElements = make(map[string]func(string, string) string)
	// Standard BBcode -> HTML tags
	b.bbElements["b"] = bbToHTML("b")
	b.bbElements["i"] = bbToHTML("i")
	b.bbElements["u"] = bbToHTML("u")
	b.bbElements["strike"] = bbToHTML("s")
	// Other BBcode tags
	b.bbElements["img"] = func(_, con string) string {
		idx := strings.IndexRune(con, '?')
		if idx > 0 {
			con = con[0:idx] + queryescape(con[idx:])
		}
		return "<a href=\"" + con + "\"><img src=\"" + con + "\" /></a>"
	}
	b.bbElements["url"] = func(par, con string) string {
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
		idx := strings.IndexRune(par, '?')
		if idx > 0 {
			par = par[0:idx] + queryescape(par[idx:])
		}
		return "<a href=\"" + par + "\" rel=\"nofollow\">" + con + "</a>"
	}
	b.bbElements["spoiler"] = func(_, con string) string {
		return "<span class=\"spoiler\">" + con + "</span>"
	}
	b.bbElements["youtube"] = func(_, con string) string {
		idx := strings.Index(con, "?v=")
		if idx > 0 {
			con = con[idx+3:]
		}
		return `<iframe width="560" height="315" src="//www.youtube.com/embed/` + url.QueryEscape(con) + `" frameborder="0" allowfullscreen></iframe>`
	}
	b.bbElements["html"] = func(_, con string) string {
		return strings.Replace(con, "\n", "", -1)
	}
	/* [video] tag accepts an optional parameter which can control
	 * the <video> params:
	 *   - nocontrols / -ctrl: disable controls on video
	 *   - autoplay / aplay: make this video autoplay
	 *   - mute / muted: mute this video
	 *   - gif / giflike: autoplay + mute + nocontrols
	 */
	b.bbElements["video"] = func(par, con string) string {
		idx := strings.LastIndex(con, ".")
		ext := con[idx+1:]
		var opts string
		if len(par) > 0 && par == "gif" {
			opts = "autoplay muted loop"
		} else {
			opts = "controls"
		}
		switch ext {
		case "webm":
			return `<video height="250px" src="` + con + `" ` + opts + `>[Your browser is unable to play this video]</video>`
		case "ogg", "ogv", "mp4":
			return `<video height="250px" ` + opts + `><source src="` + con + `" type="video/` + ext + `"/>[Your browser is unable to play this video]</video>`
		}
		return "<gray>Unsupported video type: " + ext + "</gray>"
	}
}

func (b *BBCodeFormatter) Format(code string) string {
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
					parsed := b.bbElements[tag](stack[top].Parameter, content)
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
			if _, ok := b.bbElements[tag]; ok {
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

func queryescape(query string) string {
	offset := 0
	for {
		start := index(query, offset, '=')
		if start < 0 {
			return query
		}
		end := index(query, start+1, '&')
		if end < 0 {
			return query[:start+1] + url.QueryEscape(query[start+1:])
		}
		query = query[:start+1] + url.QueryEscape(query[start+1:end]) + query[end:]
		offset = end + 1
	}
}
