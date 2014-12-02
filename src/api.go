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
	content := postContent
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
	content := postContent

	_, err = DBReplyThread(&thread, user, content)
	if err != nil {
		fmt.Println(err.Error())
		sendError(rw, 500, err.Error())
		return
	}

	http.Redirect(rw, req, "/thread/"+thread.ShortUrl+"#last", http.StatusMovedPermanently)
}

func apiEditPost(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	threadUrl := vars["thread"]
	thread, err := DBGetThread(threadUrl)
	// retreive post to edit
	postId, err := strconv.Atoi(vars["post"])
	if err != nil {
		http.Error(rw, "Invalid post ID", 400)
		return
	}
	posts, err := DBGetPosts(&thread, postId, postId)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(rw, err.Error(), 500)
		return
	}
	if len(posts) < 1 {
		http.Error(rw, "Post not found", 404)
		return
	}
	post := posts[0]
	// if post has no tripcode associated, refuse to edit
	if len(post.Author.Tripcode) < 1 {
		http.Error(rw, "Forbidden", 403);
		return
	}
	// check tripcode
	trip := req.PostFormValue("tripcode")
	if tripcode(trip) != post.Author.Tripcode {
		http.Error(rw, "Invalid tripcode", 401)
		return
	}
	// update post content and date
	newContent := req.PostFormValue("content")
	err = DBEditPost(post.Id, newContent)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}

	http.Redirect(rw, req, "/thread/" + thread.ShortUrl + "#p" + string(postId), http.StatusMovedPermanently)
}
