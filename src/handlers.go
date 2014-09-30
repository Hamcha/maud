package main

import (
	"fmt"
	"github.com/hoisie/mustache"
	"net/http"
)

func httpHome(rw http.ResponseWriter, req *http.Request) {
	send(rw, "home", nil)
}

func httpThread(rw http.ResponseWriter, req *http.Request) {

}

func httpTagSearch(rw http.ResponseWriter, req *http.Request) {

}

func httpNew(rw http.ResponseWriter, req *http.Request) {

}

func httpReply(rw http.ResponseWriter, req *http.Request) {

}

func send(rw http.ResponseWriter, name string, context interface{}) {
	fmt.Fprintf(rw,
		mustache.RenderFileInLayout(
			"template/"+name+".html",
			"template/layout.html",
			struct {
				Info SiteInfo
				Data interface{}
			}{
				siteInfo,
				context,
			}))
}
