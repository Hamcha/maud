package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type AdminRequestInfo struct {
	User     string
	BasePath string
}

var adminRequests map[*http.Request]AdminRequestInfo
var adminUsers map[string]string

func initAdmin() {
	adminRequests = make(map[*http.Request]AdminRequestInfo)
	adminUsers = viper.GetStringMapString("adminUsers")
}

func wrapAdmin(handler http.HandlerFunc, usePath bool) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, pass, _ := req.BasicAuth()
		basepath := "/"
		if usePath {
			basepath = mustGet("adminPath")
		}
		if checkAdmin(user, pass) {
			adminRequests[req] = AdminRequestInfo{
				User:     user,
				BasePath: basepath,
			}
			handler(rw, req)
			delete(adminRequests, req)
		} else {
			rw.Header().Set("WWW-Authenticate", "Basic Realm=\"maud\"")
			http.Error(rw, "Unauthorized", 401)
			return
		}
	}
}

func SetHandler(router *mux.Router, path string, handler http.HandlerFunc, isAdmin, isSubdir bool) {
	handler = wrapBlacklist(handler)
	if isAdmin {
		handler = wrapAdmin(handler, isSubdir)
	}
	router.HandleFunc(path, handler)
}

func isAdmin(req *http.Request) (bool, AdminRequestInfo) {
	if val, ok := adminRequests[req]; ok {
		return true, val
	}
	return false, AdminRequestInfo{}
}

func checkAdmin(user, pass string) bool {
	if enc, ok := adminUsers[user]; ok {
		sum := sha256.Sum256([]byte(pass))
		str := hex.EncodeToString(sum[:])
		return str == enc
	}
	return false
}
