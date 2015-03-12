package main

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gorilla/mux"
	"net/http"
)

type AdminRequestInfo struct {
	User     string
	BasePath string
}

var adminRequests map[*http.Request]AdminRequestInfo

func initAdmin() {
	adminRequests = make(map[*http.Request]AdminRequestInfo)
}

func wrapAdmin(handler http.HandlerFunc, usePath bool) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, pass, _ := req.BasicAuth()
		basepath := "/"
		if usePath {
			basepath = adminConf.Path
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

func wrapBlacklist(handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		isBanned, banReason, blAction := checkBlacklist(req)
		if isBanned {
			switch blAction {
			case "ban":
				sendError(rw, 423, banReason)
				return
			case "captcha":
				req.Header.Add("Captcha-required", "true")
			}
		}
		handler(rw, req)
		req.Header.Del("Captcha-required")
	}
}

func SetHandler(router *mux.Router, path string, handler http.HandlerFunc, isAdmin, isSubdir bool) {
	if isAdmin {
		handler = wrapAdmin(handler, isSubdir)
	}
	handler = wrapBlacklist(handler)
	router.HandleFunc(path, handler)
}

func isAdmin(req *http.Request) (bool, AdminRequestInfo) {
	if val, ok := adminRequests[req]; ok {
		return true, val
	}
	return false, AdminRequestInfo{}
}

func checkAdmin(user, pass string) bool {
	if enc, ok := adminConf.Admins[user]; ok {
		sum := sha256.Sum256([]byte(pass))
		str := hex.EncodeToString(sum[:])
		return str == enc
	}
	return false
}
