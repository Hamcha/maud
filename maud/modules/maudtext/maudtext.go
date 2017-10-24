package maudtext

import (
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/gorilla/mux"
	"github.com/hamcha/maud/maud/modules"
)

func Provide(postsPerPage int) modules.PostMutator {
	mt := maudtextMutator{postsPerPage}
	return modules.PostMutator{
		Condition: func(_ *http.Request) bool { return true },
		Mutator:   mt.maudtext,
	}
}

func init() {
	log.Printf("[ OK ] Module initialized: Maudtext")
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

		// Skip all leading HTML tags
		linestart, off, err := skipLeadingTags(line)
		if err != nil {
			continue
		}

		// find out if this is a post quote (^>>\s*#[0-9]+\s*$) or a line quote
		if line[off:off+4] == "&gt;" {
			stripped := strings.Replace(line[off:], " ", "", -1)
			stripped = strings.TrimSuffix(stripped, "<br/>")
			if len(stripped) < 10 || stripped[:9] != postQuotePrefix {
				// line quote
				if off == 0 {
					lines[idx] = string(linestart) + `<span class="purpletext">&gt; ` +
						line[off+4:] + "</span>"
				} else {
					// If it's enclosed by HTML tags, find out where is the inmost tag closing
					// and wrap the purpletext in it.
					tagclose := 0
					for i := off + 4; i < len(line); i++ {
						if line[i] == '<' {
							tagclose = i
							break
						}
					}
					if tagclose == 0 {
						log.Printf("[ ERROR ] Maudtext: Invalid HTML in line: " + line)
						continue
					}
					lines[idx] = string(linestart) + `<span class="purpletext">&gt; ` +
						line[off+4:tagclose] + "</span>" + line[tagclose:]
				}
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

// skipLeadingTags checks if the given line starts with 1 or more HTML tags and returns
// (the string of said lines, the length of that string, error), where error != nil if
// and only if the line starts with invalid HTML.
func skipLeadingTags(line string) (string, int, error) {
	linestart := make([]rune, 0)
	tagsopen := 0
skiphtmlfor:
	for _, r := range line {
		switch r {
		case '<':
			tagsopen++
		case '>':
			tagsopen--
			if tagsopen < 0 {
				log.Printf("[ ERROR ] Maudtext: Invalid HTML in line: " + line)
				return "", 0, errors.New("Invalid HTML in line")
			}
		default:
			if tagsopen <= 0 {
				break skiphtmlfor
			}
		}
		linestart = append(linestart, r)
	}
	return string(linestart), len(linestart), nil
}
