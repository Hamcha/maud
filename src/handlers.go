package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hoisie/mustache"
	"net/http"
	"strconv"
)

func httpHome(rw http.ResponseWriter, req *http.Request) {
	tags, err := DBGetPopularTags()
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	type TagData struct {
		Name       string
		LastUpdate int64
		LastThread Thread
	}

	tagdata := make([]TagData, len(tags))

	for i := range tags {
		thread, err := DBGetThreadById(tags[i].LastThread)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}

		tagdata[i] = TagData{
			Name:       tags[i].Name,
			LastUpdate: tags[i].LastUpdate,
			LastThread: thread,
		}
	}

	send(rw, "home", tagdata)
}

func httpThread(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	threadUrl := vars["thread"]
	thread, err := DBGetThread(threadUrl)
	if err != nil {
		if err.Error() == "not found" {
			sendError(rw, 404, nil)
		} else {
			sendError(rw, 500, err.Error())
		}
		return
	}

	threadPost, err := DBGetPost(thread.ThreadPost)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	posts, err := DBGetPosts(&thread, 0, 0)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}
	// Filter threadpost away
	if posts[0].Id == threadPost.Id {
		posts = posts[1:]
	}

	send(rw, "thread", struct {
		Thread     Thread
		ThreadPost Post
		Posts      []Post
	}{
		thread,
		threadPost,
		posts,
	})
}

func httpTagSearch(rw http.ResponseWriter, req *http.Request) {
	send(rw, "tagsearch", nil)
}

func httpNewThread(rw http.ResponseWriter, req *http.Request) {
	send(rw, "newthread", nil)
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

func sendError(rw http.ResponseWriter, code int, context interface{}) {
	rw.WriteHeader(code)
	fmt.Fprintf(rw,
		mustache.RenderFileInLayout(
			"errors/"+strconv.Itoa(code)+".html",
			"errors/layout.html",
			struct {
				Info SiteInfo
				Data interface{}
			}{
				siteInfo,
				context,
			}))
}
