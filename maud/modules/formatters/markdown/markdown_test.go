package markdown

import (
	"testing"

	"github.com/hamcha/maud/maud/modules"
)

const (
	inItalic            = "This string contains *italic text*."
	outItalic           = "This string contains <i>italic text</i>."
	inNonItalic         = "This string *does not contain italic."
	outNonItalic        = inNonItalic
	inEscapedItalic     = `This string contains \*escaped italic text*.`
	outEscapedItalic    = `This string contains *escaped italic text*.`
	inBold              = "This string contains **bold text**."
	outBold             = "This string contains <b>bold text</b>."
	inNonBold           = "This string ** does not contain bold."
	outNonBold          = inNonBold
	inEscapedBold       = `This string contains \*\*escaped bold text**.`
	outEscapedBold      = `This string contains **escaped bold text**.`
	inEscapedFirstBold  = `This string contains \**escaped bold text**.`
	outEscapedFirstBold = outEscapedBold
	inStroke            = "This string contains ~~stroke text~~."
	outStroke           = "This string contains <s>stroke text</s>."
	inNonStroke         = `This string contains ~~non stroked text~.`
	outNonStroke        = inNonStroke
	inEscapedStroke     = `This string contains \~~escaped stroked text~~.`
	outEscapedStroke    = `This string contains ~~escaped stroked text~~.`
	inCode              = "This string contains `some code`."
	outCode             = "This string contains <code>some code</code>."
	inNonCode           = "This string contains `no code."
	outNonCode          = inNonCode
	inEscapedCode       = "This string contains \\`escaped code`."
	outEscapedCode      = "This string contains `escaped code`."
	inBoldInItalic      = "This string contains *nested **bold** and italic*."
	outBoldInItalic     = "This string contains <i>nested <b>bold</b> and italic</i>."
	inItalicInBold      = "This string contains **nested *italic* and bold**."
	outItalicInBold     = "This string contains <b>nested <i>italic</i> and bold</b>."
	inUrl               = "This string [contains an URL](http://localhost/)"
	outUrl              = `This string <a href="http://localhost/">contains an URL</a>`
	inNonUrl            = "This string [contains no URL](http://localhost/"
	outNonUrl           = inNonUrl
	inEscapedUrl        = `This string \[contains an escaped URL](http://localhost/)`
	outEscapedUrl       = `This string [contains an escaped URL](http://localhost/)`
	inResource          = "This string ![contains a resource](http://localhost/maud.png)"
	outResource         = `This string <a href="http://localhost/maud.png" rel="nofollow"><img src="http://localhost/maud.png" alt="contains a resource"/></a>`
	inEmpty             = ""
	outEmpty            = inEmpty
)

var md modules.Formatter

func init() {
	md = Provide()
}

func shouldEq(t *testing.T, in, out string) {
	if result := md.Format(in); result != out {
		t.Error("For input\r\n\r\n    ", in, "\r\n\r\nexpected\r\n\r\n    ", out,
			"\r\n\r\nbut got\r\n\r\n    ", result)
	}
}

func TestItalic(t *testing.T) {
	shouldEq(t, inItalic, outItalic)
	shouldEq(t, inNonItalic, outNonItalic)
	shouldEq(t, inEscapedItalic, outEscapedItalic)
}

func TestBold(t *testing.T) {
	shouldEq(t, inBold, outBold)
	shouldEq(t, inNonBold, outNonBold)
	shouldEq(t, inEscapedBold, outEscapedBold)
	shouldEq(t, inEscapedFirstBold, outEscapedFirstBold)
}

func TestCode(t *testing.T) {
	shouldEq(t, inCode, outCode)
	shouldEq(t, inNonCode, outNonCode)
	shouldEq(t, inEscapedCode, outEscapedCode)
}

func TestStroke(t *testing.T) {
	shouldEq(t, inStroke, outStroke)
	shouldEq(t, inNonStroke, outNonStroke)
	shouldEq(t, inEscapedStroke, outEscapedStroke)
}

func TestBoldInItalic(t *testing.T) {
	shouldEq(t, inBoldInItalic, outBoldInItalic)
}

func TestItalicInBold(t *testing.T) {
	shouldEq(t, inItalicInBold, outItalicInBold)
}

func TestUrl(t *testing.T) {
	shouldEq(t, inUrl, outUrl)
	shouldEq(t, inNonUrl, outNonUrl)
	shouldEq(t, inEscapedUrl, outEscapedUrl)
}

func TestResource(t *testing.T) {
	shouldEq(t, inResource, outResource)
}

func TestEmpty(t *testing.T) {
	shouldEq(t, inEmpty, outEmpty)
}
