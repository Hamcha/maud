package modules

import (
	. "../data"
	"net/http"
)

type Formatter interface {
	Format(string) string
}

type PostMutatorData struct {
	Request *http.Request
	Thread  *Thread
	Post    *Post
}

type PostMutator struct {
	Condition func(*http.Request) bool
	Mutator   func(post PostMutatorData)
}
