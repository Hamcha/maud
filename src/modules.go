package main

import (
	// Formatters
	"./modules"
	"./modules/formatters/bbcode"
	"./modules/formatters/lightify"
	"./modules/formatters/markdown"
)

var formatters []modules.Formatter
var mutators []modules.Mutator

func InitFormatters() {
	formatters = make([]modules.Formatter, 0)
	mutators = make([]modules.Mutator, 0)

	// Post formatters
	formatters = append(formatters, bbcode.Provide())
	formatters = append(formatters, markdown.Provide())

	// Lightifier
	lightifier := lightify.Provide()
	lightmutator := modules.Mutator{
		Condition: isLightVersion,
		Mutator:   lightifier.ReplaceTags,
	}
	formatters = append(formatters, lightifier)
	mutators = append(mutators, lightmutator)
}
