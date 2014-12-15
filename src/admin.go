package main

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
)

var adminRequests map[*http.Request]bool

func initAdmin() {
	adminRequests = make(map[*http.Request]bool)
}

func wrapAdmin(handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		adminRequests[req] = true
		user, pass, _ := req.BasicAuth()
		if checkAdmin(user, pass) {
			handler(rw, req)
		} else {
			rw.Header().Set("WWW-Authenticate", "Basic Realm=\"maud\"")
			http.Error(rw, "Unauthorized", 401)
			return
		}
		delete(adminRequests, req)
	}
}

func SetHandler(router *mux.Router, path string, handler http.HandlerFunc, isAdmin bool) {
	if isAdmin {
		handler = wrapAdmin(handler)
	}

	router.HandleFunc(path, handler)
}

func isAdmin(req *http.Request) bool {
	if val, ok := adminRequests[req]; ok {
		return val
	}
	return false
}

func checkAdmin(user, pass string) bool {
	if enc, ok := adminConf.Admins[user]; ok {
		sum := sha256.Sum256([]byte(pass))
		str := hex.EncodeToString(sum[:])
		return str == enc
	}
	return false
}
