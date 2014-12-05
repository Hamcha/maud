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
	p.AllowElements("b", "i", "pre", "small", "strike", "tt", "u")
	p.AllowElements("blockquote", "cite")
	p.AllowElements("br", "div", "hr", "p", "span", "wbr")

	// Links
	p.AllowAttrs("href").OnElements("a")

	// Extra markup
	p.AllowElements("abbr", "acronym", "cite", "code", "dfn", "em",
		"figcaption", "mark", "s", "samp", "strong", "sub", "sup", "var")

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
