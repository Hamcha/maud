package main

import (
	"github.com/microcosm-cc/bluemonday"
)

func PostPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()

	p.AllowStandardAttributes()
	p.AllowStandardURLs()

	// Basic markup
	p.AllowElements("h1", "h2", "h3", "h4", "h5", "h6")
	p.AllowElements("b", "em", "i", "pre", "s", "small", "u")
	p.AllowElements("blockquote", "cite", "q")
	p.AllowElements("br", "div", "hr", "p", "span", "wbr")

	// Links
	p.AllowAttrs("href").OnElements("a")

	// Extra markup
	p.AllowElements("abbr", "acronym", "bdi", "code", "dd", "del",
		"dfn", "dl", "dt", "figcaption", "figure", "ins", "kbd", "mark",
		"meter", "ol", "progress", "samp", "strong", "sub",
		"sup", "time", "ul", "var")

	p.AllowAttrs("value", "min", "max").OnElements("meter", "progress")

	// Structures
	p.AllowLists()
	p.AllowTables()

	// Multimedia
	p.AllowImages()
	p.AllowElements("video", "audio", "source", "svg", "iframe")
	p.AllowAttrs("width", "height", "src", "frameborder", "allowfullscreen").OnElements("iframe")

	p.RequireParseableURLs(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)
	p.RequireNoFollowOnFullyQualifiedLinks(true)

	return p
}
