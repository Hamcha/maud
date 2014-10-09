package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
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
		fmt.Println(err.Error())
		sendError(rw, 500, err.Error())
		return
	}

	http.Redirect(rw, req, "/thread/"+threadId, http.StatusMovedPermanently)
}

func apiReply(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	threadUrl := vars["thread"]
	thread, err := DBGetThread(threadUrl)

	postNickname := req.PostFormValue("nickname")
	postContent := req.PostFormValue("text")
	if len(postContent) < 1 {
		http.Error(rw, "Required fields are missing", 400)
		return
	}

	nickname, tripcode := parseNickname(postNickname)
	user := User{nickname, tripcode}
	content := parseContent(postContent)

	msgId, err := DBReplyThread(&thread, user, content)
	if err != nil {
		fmt.Println(err.Error())
		sendError(rw, 500, err.Error())
		return
	}

	http.Redirect(rw, req, "/thread/"+thread.ShortUrl+"#p"+strconv.Itoa(msgId-2), http.StatusMovedPermanently)
}
