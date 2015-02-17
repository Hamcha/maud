package modules

import (
	"net/http"
)

type Formatter interface {
	Format(string) string
}

type PostMutatorData struct {
	Request *http.Request
	Content *string
}

type PostMutator struct {
	Condition func(*http.Request) bool
	Mutator   func(post PostMutatorData)
}
