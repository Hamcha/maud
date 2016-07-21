package maudtext

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	modules ".."
	"github.com/gorilla/mux"
)

func Provide(postsPerPage int) modules.PostMutator {
	mt := maudtextMutator{postsPerPage}
	return modules.PostMutator{
		Condition: func(_ *http.Request) bool { return true },
		Mutator:   mt.maudtext,
	}
}

type maudtextMutator struct {
	postsPerPage int
}

// converts:
//   >> #postId
//   > quote
func (mt *maudtextMutator) maudtext(data modules.PostMutatorData) {
	lines := strings.Split((*data.Post).Content, "\n")
	vars := mux.Vars(data.Request)
	threadUrl, threadok := vars["thread"]

	const postQuotePrefix = "&gt;&gt;#"
	// Accept whichever number of spaces between >> and # (but none between # and digits!)
	pqPrefixRgx := regexp.MustCompile(`&gt;&gt;\s*#`)

	for idx := range lines {
		line := strings.TrimSpace(lines[idx])
		if len(line) < 5 {
			continue
		}
		stripped := strings.Replace(line, " ", "", -1)
		stripped = strings.TrimSuffix(stripped, "<br/>")
		if line[:4] == "&gt;" {
			// find out if this is a post quote (^>>\s*#[0-9]+\s*$) or a line quote
			if len(stripped) < 10 || stripped[:9] != postQuotePrefix {
				// line quote
				lines[idx] = "<span class=\"purpletext\">&gt; " + line[4:] + "</span>"
				continue
			}
		}
		if !threadok {
			// can only insert post quotes if thread is valid
			continue
		}

		// First, split string by the '>>#' delimiter
		split := pqPrefixRgx.Split(line, -1)
		if len(split) < 2 {
			continue
		}
		textBefore := split[0]
		// Drop the first element, as it's before the '>>#'
		split = split[1:]
		out := ""
		for _, s := range split {
			idx := 0
			// If this string starts with a digit, it was a valid post quote
			for i, ch := range s {
				if !unicode.IsDigit(ch) {
					break
				}
				idx = i
			}
			idx++
			num, err := strconv.ParseInt(s[:idx], 10, 32)
			if err != nil {
				out += `&gt;&gt; #` + s
				continue
			}
			// Insert the link, plus all extra characters we found after the digits, if any
			out += `<a href="` + mt.getLink(int(num), threadUrl) +
				`" class="postIdQuote">&gt;&gt; #` + s[:idx] + `</a>` + s[idx:]
		}
		lines[idx] = textBefore + out
	}

	(*data.Post).Content = strings.Join(lines, "\n")
}

func (mt *maudtextMutator) getLink(postNum int, threadUrl string) string {
	if postNum == 0 {
		return "/thread/" + threadUrl + "/page/1#thread"
	}
	page := postNum/mt.postsPerPage + 1
	return "/thread/" + threadUrl + "/page/" + strconv.Itoa(page) + "#p" + strconv.Itoa(postNum)
}
