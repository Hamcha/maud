package maudtext

import (
	"strconv"
	"strings"
	"testing"
)

const (
	gt    = "&gt;"
	ptPfx = `<span class="purpletext">` + gt + ` `
	ptSfx = `</span>`
)

const (
	inJustGt              = gt
	outJustGt             = ptPfx + ptSfx
	inSingleLineQuote     = gt + ` using TDD`
	outSingleLineQuote    = ptPfx + ` using TDD` + ptSfx
	inSingleLineCyrillic  = gt + ` implying I know what Стрелка means`
	outSingleLineCyrillic = ptPfx + ` implying I know what Стрелка means` + ptSfx
	inSingleLineJap       = gt + ` 日本語`
	outSingleLineJap      = ptPfx + ` 日本語` + ptSfx
	inSingleLineQuoteWs   = `   ` + gt + `using TDD   `
	outSingleLineQuoteWs  = ptPfx + `using TDD` + ptSfx
	inMultiLineQuote      = ` ` + gt + `using TDD
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
	outQuoteInTag     = ptPfx + `<strong>using TDD</strong>` + ptSfx
	inQuoteInTagWs    = `<strong>   ` + gt + `using TDD</strong>`
	outQuoteInTagWs   = ptPfx + `<strong>using TDD</strong>` + ptSfx
	inQuoteInTags     = `<strong><em><p>` + gt + `using TDD` + `</p></em></strong>`
	outQuoteInTags    = ptPfx + `<strong><em><p>using TDD</p></em></strong>` + ptSfx
	inNonPostQuote    = gt + gt + ` # 42`
	outNonPostQuote   = ptPfx + gt + ` # 42` + ptSfx

	inSingleLinePostQuote          = gt + gt + ` #42`
	inSingleLineMultiplePostQuotes = gt + gt + `#33 bla bla 日本語 ` + gt + gt + ` #22 bla.`
	inSingleLinePostQuoteWs        = `  ` + gt + gt + `   #42  `
	inSingleLinePostQuoteTags      = `<em><span>  ` + gt + gt + ` #1</span>  ` + gt + gt + `#101</em>`
	inMultiLinePostQuote           = gt + ` this is not a postquote ` + gt + gt + `#33
but this is: ` + gt + gt + `#2 and also ` + gt + gt + `    #555 this
and this is ` + gt + ` not a line quote.`
)

var (
	outSingleLinePostQuote          = pq(42)
	outSingleLineMultiplePostQuotes = pq(33) + ` bla bla 日本語 ` + pq(22) + ` bla.`
	outSingleLinePostQuoteWs        = pq(42)
	outSingleLinePostQuoteTags      = `<em><span>  ` + pq(1) + `</span>  ` + pq(101) + `</em>`
	outMultiLinePostQuote           = ptPfx + ` this is not a postquote ` + gt + gt + `#33` + ptSfx + `
but this is: ` + pq(2) + ` and also ` + pq(555) + ` this
and this is ` + gt + ` not a line quote.`
)

func pq(postNum int) string {
	var page string
	if postNum == 0 {
		page = "1#thread"
	} else {
		page = strconv.Itoa(postNum/postsPerPage+1) + `#p` + strconv.Itoa(postNum)
	}
	return `<a href="/thread/test/page/` + page + `" class="postIdQuote">&gt;&gt; #` +
		strconv.Itoa(postNum) + `</a>`
}

const postsPerPage = 20

var mt maudtextMutator

func init() {
	mt = maudtextMutator{postsPerPage, ""}
}

func TestMaudtext(t *testing.T) {
	shouldMutate(t, inJustGt, outJustGt)
	shouldMutate(t, inSingleLineQuote, outSingleLineQuote)
	shouldMutate(t, inSingleLineQuoteWs, outSingleLineQuoteWs)
	shouldMutate(t, inSingleLineCyrillic, outSingleLineCyrillic)
	shouldMutate(t, inSingleLineJap, outSingleLineJap)
	shouldMutate(t, inMultiLineQuote, outMultiLineQuote)
	shouldMutate(t, inNonQuote, outNonQuote)
	shouldMutate(t, inMultipleQuotes, outMultipleQuotes)
	shouldMutate(t, inQuoteInTag, outQuoteInTag)
	shouldMutate(t, inQuoteInTagWs, outQuoteInTagWs)
	shouldMutate(t, inQuoteInTags, outQuoteInTags)
	shouldMutate(t, inNonPostQuote, outNonPostQuote)
	shouldMutate(t, inSingleLinePostQuote, outSingleLinePostQuote)
	shouldMutate(t, inSingleLineMultiplePostQuotes, outSingleLineMultiplePostQuotes)
	shouldMutate(t, inSingleLinePostQuoteWs, outSingleLinePostQuoteWs)
	shouldMutate(t, inSingleLinePostQuoteTags, outSingleLinePostQuoteTags)
	shouldMutate(t, inMultiLinePostQuote, outMultiLinePostQuote)
}

func shouldMutate(t *testing.T, in, out string) {
	// We mock the actual Mutator function, rather than using it, as it requires
	// a valid http.Request. Since the involved logic is so easy, it's more
	// convenient to just replicate it instead of reconstructing a fully valid Request.
	mt.threadUrl = "test"
	lines := strings.Split(in, "\n")
	for idx := range lines {
		lines[idx] = mt.maudtext(strings.TrimSpace(lines[idx]))
	}
	result := strings.Join(lines, "\n")
	if result != out {
		t.Error("Maudtext: for input\r\n\r\n    ", "^"+in+
			"$", "\r\n\r\nexpected\r\n\r\n    ", "^"+out+"$",
			"\r\n\r\nbut got\r\n\r\n    ", "^"+result+"$")
	}
}
