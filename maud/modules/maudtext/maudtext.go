package maudtext

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hamcha/maud/maud/modules"
)

func Provide(postsPerPage int) modules.PostMutator {
	mt := maudtextMutator{postsPerPage, ""}
	return modules.PostMutator{
		Condition: func(_ *http.Request) bool { return true },
		Mutator:   mt.mutate,
	}
}

func init() {
	log.Printf("[ OK ] Module initialized: Maudtext")
}

type maudtextMutator struct {
	postsPerPage int
	threadUrl    string
}

// converts:
//   >> #postId
//   > quote
func (mt *maudtextMutator) mutate(data modules.PostMutatorData) {
	lines := strings.Split((*data.Post).Content, "\n")
	vars := mux.Vars(data.Request)
	threadUrl := vars["thread"]
	mt.threadUrl = threadUrl

	for idx := range lines {
		lines[idx] = mt.maudtext(strings.TrimSpace(lines[idx]))
	}

	(*data.Post).Content = strings.Join(lines, "\n")
}

func (mt *maudtextMutator) getLink(postNum int) string {
	if postNum == 0 {
		return "/thread/" + mt.threadUrl + "/page/1#thread"
	}
	page := postNum/mt.postsPerPage + 1
	return "/thread/" + mt.threadUrl + "/page/" + strconv.Itoa(page) + "#p" + strconv.Itoa(postNum)
}
