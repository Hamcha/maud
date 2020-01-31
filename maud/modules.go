package main

import (
	. "github.com/hamcha/maud/maud/data"
	"github.com/hamcha/maud/maud/modules"
	"github.com/spf13/viper"

	// Formatters
	"github.com/hamcha/maud/maud/modules/formatters/bbcode"
	"github.com/hamcha/maud/maud/modules/formatters/lightify"
	"github.com/hamcha/maud/maud/modules/formatters/markdown"

	// Mutators
	"github.com/hamcha/maud/maud/modules/maudtext"
	"github.com/hamcha/maud/maud/modules/proxy"

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

	// Lightifier
	lightifier := lightify.Provide()
	lightmutator := modules.PostMutator{
		Condition: isLightVersion,
		Mutator:   lightifier.ReplaceTags,
	}
	formatters = append(formatters, lightifier)
	postmutators = append(postmutators, lightmutator)

	// Maudtext (handles quotes etc)
	postmutators = append(postmutators, maudtext.Provide(viper.GetInt("postsPerPage")))

	// Proxy
	if viper.GetBool("useProxy") {
		postmutators = append(postmutators, proxy.Provide(mustGet("proxyRoot"), mustGet("proxyDomain")))
	}
}

func applyPostMutator(m modules.PostMutator, thread *Thread, post *Post, rw *http.ResponseWriter, req *http.Request) {
	if m.Condition(req) {
		m.Mutator(postMutatorArgs(thread, post, rw, req))
	}
}

func postMutatorArgs(thread *Thread, post *Post, rw *http.ResponseWriter, req *http.Request) modules.PostMutatorData {
	return modules.PostMutatorData{
		Thread:         thread,
		Post:           post,
		Request:        req,
		ResponseWriter: rw,
	}
}
