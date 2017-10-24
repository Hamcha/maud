package maudtext

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"

	. "github.com/hamcha/maud/maud/data"
	"github.com/hamcha/maud/maud/modules"
)

const (
	gt    = "&gt;"
	ptPfx = `<span class="purpletext">` + gt + ` `
	ptSfx = `</span>`
	pqSfx = `</a>`
)

const (
	inJustGt             = gt
	outJustGt            = gt
	inSingleLineQuote    = gt + ` using TDD`
	outSingleLineQuote   = ptPfx + ` using TDD` + ptSfx
	inSingleLineQuoteWs  = `   ` + gt + `using TDD   `
	outSingleLineQuoteWs = ptPfx + `using TDD` + ptSfx
	inMultiLineQuote     = ` ` + gt + `using TDD
test

` + gt + ` implying this will avoid bugs
`
	outMultiLineQuote = ptPfx + `using TDD` + ptSfx + `
test

` + ptPfx + ` implying this will avoid bugs` + ptSfx + `
`
	inNonQuote        = `three ` + gt + ` two`
	outNonQuote       = inNonQuote
	inMultipleQuotes  = gt + gt + gt + gt + ` using TDD`
	outMultipleQuotes = ptPfx + gt + gt + gt + ` using TDD` + ptSfx
	inQuoteInTag      = `<strong>` + gt + `using TDD</strong>`
	outQuoteInTag     = `<strong>` + ptPfx + `using TDD` + ptSfx + `</strong>`
	inQuoteInTags     = `<strong><em><p>` + gt + `using TDD` + `</p></em></strong>`
	outQuoteInTags    = `<strong><em><p>` + ptPfx + `using TDD` + ptSfx + `</p></em></strong>`
	inNonPostQuote    = gt + gt + ` # 42`
	outNonPostQuote   = ptPfx + gt + ` # 42` + ptSfx

	inSingleLinePostQuote = gt + gt + ` #42`
)

var (
	outSingleLinePostQuote = pqPfx(42) + pqSfx
)

func pqPfx(postNum int) string {
	var page string
	if postNum == 0 {
		page = "1#thread"
	} else {
		page = strconv.Itoa(postNum/postsPerPage+1) + `#p` + strconv.Itoa(postNum)
	}
	return `<a href="/thread/test/page/` + page + `" class="postIdQuote">&gt;&gt; #` +
		strconv.Itoa(postNum)
}

const postsPerPage = 20

var mt modules.PostMutator
var (
	post   Post
	thread Thread
	req    *http.Request
)

func init() {
	mt = Provide(postsPerPage)
	thread.ShortUrl = "test"
	url, _ := url.Parse("https://crunchy.rocks/thread/test")
	req = &http.Request{}
	req.URL = url
}

func TestMaudtext(t *testing.T) {
	shouldMutate(t, inJustGt, outJustGt)
	shouldMutate(t, inSingleLineQuote, outSingleLineQuote)
	shouldMutate(t, inSingleLineQuoteWs, outSingleLineQuoteWs)
	shouldMutate(t, inMultiLineQuote, outMultiLineQuote)
	shouldMutate(t, inNonQuote, outNonQuote)
	shouldMutate(t, inMultipleQuotes, outMultipleQuotes)
	shouldMutate(t, inQuoteInTag, outQuoteInTag)
	shouldMutate(t, inQuoteInTags, outQuoteInTags)
	//shouldMutate(t, inNonPostQuote, outNonPostQuote)
	// FIXME: post quoting is non trivial to unit-test, as it calls
	// mux.Vars internally.
	//shouldMutate(t, inSingleLinePostQuote, outSingleLinePostQuote)
}

func shouldMutate(t *testing.T, in, out string) {
	post.Content = in
	pmd := modules.PostMutatorData{
		Request:        req,
		ResponseWriter: nil,
		Thread:         &thread,
		Post:           &post,
	}
	// Mutator does not return a value, but mutates in-place.
	mt.Mutator(pmd)
	if result := (*pmd.Post).Content; result != out {
		t.Error("Maudtext: for input\r\n\r\n    ", "^"+in+
			"$", "\r\n\r\nexpected\r\n\r\n    ", "^"+out+"$",
			"\r\n\r\nbut got\r\n\r\n    ", "^"+result+"$")
	}
}
