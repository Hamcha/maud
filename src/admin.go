package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

var adminRequests map[*http.Request]bool

func initAdmin() {
	adminRequests = make(map[*http.Request]bool)
}

func wrapAdmin(handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		adminRequests[req] = true
		handler(rw, req)
	}
}

func SetHandler(router *mux.Router, path string, handler http.HandlerFunc, isAdmin bool) {
	if isAdmin {
		handler = wrapAdmin(handler)
	}

	router.HandleFunc(path, handler)
}
