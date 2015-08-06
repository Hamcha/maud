package maudtext

import (
	".."
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

func Provide(postsPerPage int) modules.PostMutator {
	mt := MaudtextMutator{postsPerPage}
	return modules.PostMutator{
		Condition: func(_ *http.Request) bool { return true },
		Mutator:   mt.maudtext,
	}
}

type MaudtextMutator struct {
	postsPerPage int
}

// converts:
//   >> #postId
//   > quote
func (mt *MaudtextMutator) maudtext(data modules.PostMutatorData) {
	lines := strings.Split((*data.Post).Content, "\n")
	vars := mux.Vars(data.Request)
	threadUrl, threadok := vars["thread"]

	for idx := range lines {
		line := strings.TrimSpace(lines[idx])
		if len(line) < 5 || line[:4] != "&gt;" {
			continue
		}
		// find out if this is a post quote (^>>\s*#[0-9]+\s*$) or a line quote
		stripped := strings.Replace(line, " ", "", -1)
		stripped = strings.TrimSuffix(stripped, "<br/>")
		if threadok && len(stripped) > 9 && stripped[:9] == "&gt;&gt;#" {
			if num, err := strconv.ParseInt(stripped[9:], 10, 32); err == nil {
				// valid post quote
				lines[idx] = "<a href=\"" + mt.getLink(int(num), threadUrl) +
					"\" class=\"postIdQuote\">&gt;&gt; #" + strconv.Itoa(int(num)) + "</a><br/>"
				continue
			}
		}
		lines[idx] = "<span class=\"purpletext\">&gt; " + line[4:] + "</span>"
	}

	(*data.Post).Content = strings.Join(lines, "\n")
}

func (mt *MaudtextMutator) getLink(postNum int, threadUrl string) string {
	if postNum == 0 {
		return "/thread/" + threadUrl + "/page/1#thread"
	}
	page := postNum/mt.postsPerPage + 1
	return "/thread/" + threadUrl + "/page/" + strconv.Itoa(page) + "#p" + strconv.Itoa(postNum)
}
