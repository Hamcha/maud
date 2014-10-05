package main

import (
	"net/http"
)

func apiNewThread(rw http.ResponseWriter, req *http.Request) {
	postTitle := req.PostFormValue("title")
	postNickname := req.PostFormValue("nickname")
	postContent := req.PostFormValue("text")
	postTags := req.PostFormValue("tags")
	if len(postTitle) < 1 || len(postContent) < 1 {
		http.Error(rw, "Required fields are missing", 400)
		return
	}

	nickname, tripcode := parseNickname(postNickname)
	user := User{nickname, tripcode}
	content := parseContent(postContent)
	tags := parseTags(postTags)

	threadId, err := DBNewThread(user, postTitle, content, tags)
	if err != nil {
		http.Error(rw, "Something went wrong...", 500)
	}

	http.Redirect(rw, req, "/thread/"+threadId, http.StatusMovedPermanently)
}

func apiReply(rw http.ResponseWriter, req *http.Request) {
	http.Error(rw, "MY HAMON IS NOT STRONG ENOUGH YET", 501)
}
