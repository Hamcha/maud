package main

import (
	. "./data"
	// Formatters
	"./modules"
	"./modules/formatters/bbcode"
	"./modules/formatters/lightify"
	"./modules/formatters/markdown"
	"./modules/formatters/proxy"
	// Mutators
	"./modules/maudtext"
	// Go libs
	"net/http"
)

var formatters []modules.Formatter
var postmutators []modules.PostMutator

func InitFormatters() {
	formatters = make([]modules.Formatter, 0)
	postmutators = make([]modules.PostMutator, 0)

	// Post formatters
	formatters = append(formatters, bbcode.Provide())
	formatters = append(formatters, markdown.Provide())
	if siteInfo.UseProxy {
		formatters = append(formatters, proxy.Provide(siteInfo.ProxyRoot))
	}

	// Lightifier
	lightifier := lightify.Provide()
	lightmutator := modules.PostMutator{
		Condition: isLightVersion,
		Mutator:   lightifier.ReplaceTags,
	}
	formatters = append(formatters, lightifier)
	postmutators = append(postmutators, lightmutator)

	// Maudtext (handles quotes etc)
	postmutators = append(postmutators, maudtext.Provide(siteInfo.PostsPerPage))
}

func applyPostMutator(m modules.PostMutator, thread *Thread, post *Post, req *http.Request) {
	if m.Condition(req) {
		m.Mutator(postMutatorArgs(thread, post, req))
	}
}

func postMutatorArgs(thread *Thread, post *Post, req *http.Request) modules.PostMutatorData {
	return modules.PostMutatorData{
		Thread:  thread,
		Post:    post,
		Request: req,
	}
}
