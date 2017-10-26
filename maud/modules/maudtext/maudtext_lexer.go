package maudtext

/* Maudtext matcher. Processes lines matching:
 *
 * &gt; line
 * <tags>&gt; line</tags>
 * &gt;&gt; #num [...] &gt;&gt; #num2 [etc]
 * <tags>&gt;&gt; #num [...] &gt;&gt; #num2 [etc]
 *
 *
 * The matcher works as a FSM, where the logic goes like this:
 *
 * [START] -> (match Start Line) -> <matched line quote> -> END
 *                               -> <matched post quote> -> (match Post Quote) until EOL -> END
 *                               -> <matched neither>   -/
 */

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type lexState int

// A stateFunc takes an input string and a buffer of the runes processed so far.
// It returns the remaining unprocessed string, the updated processed buffer
// and the next state.
type stateFunc func(string, int) (string, int, lexState)
type multiCasePredicate func(rune) (bool, lexState)

const (
	// Start Line states
	stateSLStart lexState = iota
	stateSLAmp1
	stateSLG1
	stateSLT1
	stateSLSemi1
	stateSLAmp2
	stateSLG2
	stateSLT2
	stateSLSemi2
	stateSLPound
	stateSLFirstDigit
	stateSLOpenTag

	// Post Quote states
	statePQStart
	statePQAmp1
	statePQG1
	statePQT1
	statePQSemi1
	statePQAmp2
	statePQG2
	statePQT2
	statePQSemi2
	statePQPound
	statePQFirstDigit

	// Terminals
	stateFail
	stateLineQuote
	statePostQuote
)

func (s lexState) isTerminal() bool {
	return s == stateFail || s == stateLineQuote || s == statePostQuote
}

type stateMachine map[lexState]stateFunc

///// State machines definition

// startLineSM can end with states:
// - stateFail (didn't match a leading &gt;)
// - stateLineQuote (did match a leading &gt; but not a leading &gt;&gt;#num)
// - statePostQuote (did match a leading &gt;&gt;#num)
var startLineSM = stateMachine{
	// Start Line
	stateSLStart: makeMultiCaseLexer([]multiCasePredicate{
		func(r rune) (bool, lexState) { return r == '&', stateSLAmp1 },
		func(r rune) (bool, lexState) { return r == '<', stateSLOpenTag },
		func(r rune) (bool, lexState) { return unicode.IsSpace(r), stateSLStart },
	}, stateFail),
	stateSLAmp1:  makeSingleRuneLexer('g', stateSLG1, stateFail),
	stateSLG1:    makeSingleRuneLexer('t', stateSLT1, stateFail),
	stateSLT1:    makeSingleRuneLexer(';', stateSLSemi1, stateFail),
	stateSLSemi1: makeSingleRuneLexer('&', stateSLAmp2, stateLineQuote),
	stateSLAmp2:  makeSingleRuneLexer('g', stateSLG2, stateLineQuote),
	stateSLG2:    makeSingleRuneLexer('t', stateSLT2, stateLineQuote),
	stateSLT2:    makeSingleRuneLexer(';', stateSLSemi2, stateLineQuote),
	stateSLSemi2: makeMultiCaseLexer([]multiCasePredicate{
		func(r rune) (bool, lexState) { return r == '#', stateSLPound },
		func(r rune) (bool, lexState) { return unicode.IsSpace(r), stateSLSemi2 },
	}, stateLineQuote),
	stateSLPound:      makeDigitLexer(stateSLFirstDigit, stateLineQuote),
	stateSLFirstDigit: makeDigitLexer(stateSLFirstDigit, statePostQuote),
	stateSLOpenTag: makeMultiCaseLexer([]multiCasePredicate{
		func(r rune) (bool, lexState) { return r == '>', stateSLStart },
		func(r rune) (bool, lexState) { return true, stateSLOpenTag },
	}, stateFail),
}

// postQuoteSM can end with states:
// - stateFail (didn't match a &gt;&gt;#num)
// - statePostQuote (did match it)
var postQuoteSM = stateMachine{
	// Post Quote
	statePQStart: makeMultiCaseLexer([]multiCasePredicate{
		func(r rune) (bool, lexState) { return r == '&', statePQAmp1 },
		func(r rune) (bool, lexState) { return unicode.IsSpace(r), statePQStart },
		func(r rune) (bool, lexState) { return true, statePQStart },
	}, stateFail),
	statePQAmp1:  makeSingleRuneLexer('g', statePQG1, statePQStart),
	statePQG1:    makeSingleRuneLexer('t', statePQT1, statePQStart),
	statePQT1:    makeSingleRuneLexer(';', statePQSemi1, statePQStart),
	statePQSemi1: makeSingleRuneLexer('&', statePQAmp2, statePQStart),
	statePQAmp2:  makeSingleRuneLexer('g', statePQG2, statePQStart),
	statePQG2:    makeSingleRuneLexer('t', statePQT2, statePQStart),
	statePQT2:    makeSingleRuneLexer(';', statePQSemi2, statePQStart),
	statePQSemi2: makeMultiCaseLexer([]multiCasePredicate{
		func(r rune) (bool, lexState) { return r == '#', statePQPound },
		func(r rune) (bool, lexState) { return unicode.IsSpace(r), statePQSemi2 },
	}, statePQStart),
	statePQPound:      makeDigitLexer(statePQFirstDigit, statePQStart),
	statePQFirstDigit: makeDigitLexer(statePQFirstDigit, statePostQuote),
}

// Run makes the SM process the input and returns
// (unprocessed string, processed string, final state)
func (sm *stateMachine) Run(in string, state lexState) (string, string, lexState) {
	//buf := make(int, 0, 1<<12)
	cursor := 0
	s := in
	for !state.isTerminal() {
		f := (*sm)[state]
		s, cursor, state = f(s, cursor)
	}
	return s, in[:cursor], state
}

func (mt *maudtextMutator) maudtext(in string) string {
	// Run the StartLine SM
	rem, proc, state := startLineSM.Run(in, stateSLStart)
	if state == stateLineQuote {
		return handleLineQuote(rem, proc)
	}

	if len(mt.threadUrl) == 0 {
		// can only insert post quotes if thread is valid
		return in
	}

	var processed string

	if state == statePostQuote {
		processed, rem = mt.handlePostQuote(rem, proc)
	}

	// If necessary, lex the rest of the line
	for len(rem) > 0 {
		in = rem
		rem, proc, state = postQuoteSM.Run(rem, statePQStart)
		if state == statePostQuote {
			var p string
			p, rem = mt.handlePostQuote(rem, proc)
			processed += p
		} else {
			processed += proc
		}
	}

	return processed
}

func wrapPurpleText(s string) string {
	return `<span class="purpletext">&gt; ` + s + `</span>`
}

func handleLineQuote(rem, proc string) string {
	// rem contains the "remaining" string, proc contains the matched '>' (possibly preceded by tags)
	// We sometimes need to backtrack, as we may have been called in a situation where
	// proc = "&gt;&gt", rem = "stuff"
	// In this case we must rewind `proc` to the first instance of "&gt;" (after leading tags) and
	// merge the trailing runes into `rem`.

	// Easy case: proc contains the line up to '>', rem the rest
	if len(proc) == len(`&gt;`) {
		// No leading HTML tags
		return wrapPurpleText(rem)
	}

	// Isolate leading tags
	var leading string
	rest := proc
	lasttagidx := strings.LastIndexByte(proc, '>')
	if lasttagidx >= 0 {
		leading = proc[:lasttagidx+1]
		rest = proc[lasttagidx+1:]
	}
	rest = strings.TrimSpace(rest)

	// The actual quoted line is everything after the first &gt;
	qlidx := strings.IndexByte(rest, ';')

	return wrapPurpleText(leading + rest[qlidx+1:] + rem)
}

// handleFirstPostQuote formats the first >>#postquote of the line, taking leading tags into
// account. It returns (the formatted post quote, the rest of the string)
func (mt *maudtextMutator) handlePostQuote(rem, proc string) (string, string) {
	// rem contains the "remaining" string, proc contains a string ending with the matched postquote

	var leading string
	if proc[0] != '&' {
		// We look for the last occurrence of ';', which is known to end the sequence "&gt;&gt;".
		// Everything before that sequence is the leading string.
		lastsemiidx := strings.LastIndexByte(proc, ';')
		leading = proc[:lastsemiidx-len(`&gt;&gt`)]
	}

	poundidx := strings.LastIndexByte(proc, '#')
	digits := proc[poundidx+1:]
	n := 0
	for _, r := range digits {
		n = 10*n + int(r-'0')
	}
	return leading + `<a href="` + mt.getLink(n) +
		`" class="postIdQuote">&gt;&gt; #` + digits + `</a>`, rem
}

func makeMultiCaseLexer(predicates []multiCasePredicate, stateFail lexState) stateFunc {
	return func(in string, cursor int) (string, int, lexState) {
		for i, r := range in {
			for _, pred := range predicates {
				if ok, newstate := pred(r); ok {
					_, length := utf8.DecodeRuneInString(in[i:])
					cursor += i + length
					return in[i+length:], cursor, newstate
				}
			}
			return in, cursor, stateFail
		}
		return in, cursor, stateFail
	}
}

func makeSingleRuneLexer(match rune, stateMatch, stateFail lexState) stateFunc {
	return func(in string, cursor int) (string, int, lexState) {
	out:
		for i, r := range in {
			switch r {
			case match:
				_, length := utf8.DecodeRuneInString(in[i:])
				cursor += i + length
				return in[i+length:], cursor, stateMatch
			default:
				break out
			}
		}
		return in, cursor, stateFail
	}
}

func makeDigitLexer(stateMatch, stateFail lexState) stateFunc {
	return func(in string, cursor int) (string, int, lexState) {
	out:
		for i, r := range in {
			switch {
			case unicode.IsDigit(r):
				cursor += i + 1 // A digit will always be of length 1
				return in[i+1:], cursor, stateMatch
			default:
				break out
			}
		}
		return in, cursor, stateFail
	}
}
