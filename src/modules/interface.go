package modules

import (
	"net/http"
)

type Formatter interface {
	Format(string) string
}

type Mutator struct {
	Condition func(*http.Request) bool
	Mutator   func(content, threadUrl string) string
}
