package main

import (
	"./data"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
)

// apiNewThread: creates a new thread with its OP.
// POST params: title, text, [nickname, tags]
func apiNewThread(rw http.ResponseWriter, req *http.Request) {
	// check if this IP is blacklisted
	if isBanned, banReason, blAction := checkBlacklist(req); isBanned {
		switch blAction {
		case "ban":
			sendError(rw, 423, banReason)
			return
		case "captcha":
			// require captcha parameter
			postUserCaptcha := strings.Replace(req.PostFormValue("captcha"), " ", "", -1)
			postUserCaptcha = strings.ToLower(postUserCaptcha)
			postCaptchaQuestion := req.PostFormValue("question")
			found := false
			postRealCaptcha := ""
			for idx := range captchas {
				if captchas[idx].Question == postCaptchaQuestion {
					postRealCaptcha = captchas[idx].Answer
					found = true
					break
				}
			}
			if !found || postUserCaptcha != postRealCaptcha {
				sendError(rw, 401, "Incorrect captcha.")
				return
			}
		}
	}
	postTitle := req.PostFormValue("title")
	postNickname := req.PostFormValue("nickname")
	postContent := strings.TrimSpace(req.PostFormValue("text"))
	postTags := req.PostFormValue("tags")
	if len(postTitle) < 1 || len(postContent) < 1 {
		sendError(rw, 400, "Required fields are missing")
		return
	}

	postAntispam := req.PostFormValue("ferrarelle")
	if len(postAntispam) > 0 {
		// Fuck spambots
		sendError(rw, 500, "Database error")
		return
	}

	nickname, tcode := parseNickname(postNickname)
	var hTrip string
	if len(tcode) < 1 {
		hTrip = randomString(8)
		tcode = tripcode(hTrip)
	}
	user := data.User{nickname, tcode, len(hTrip) > 0}
	content := postContent
	tags := parseTags(postTags)

	if postTooLong(content) {
		sendError(rw, 400, "Post is too long.")
		return
	}

	threadId, err := db.NewThread(user, postTitle, content, tags)
	if err != nil {
		fmt.Println(err.Error())
		debug.PrintStack()
		sendError(rw, 500, "Database error: "+err.Error())
		return
	}

	basepath := "/"
	if ok, val := isAdmin(req); ok {
		basepath = val.BasePath
	}

	if len(hTrip) > 0 {
		http.SetCookie(rw, &http.Cookie{
			Name:     "crSetLatestPost",
			Value:    threadId + "/0/" + hTrip,
			Path:     "/thread/",
			MaxAge:   600,
			HttpOnly: false,
		})
	}
	http.Redirect(rw, req, basepath+"thread/"+threadId, http.StatusMovedPermanently)
}

// apiReply: appends a post to a thread. If POST parameter 'nickname' is given
// and has a tripcode, use that as "visible tripcode", else generate a 'hidden
// tripcode' and use that as tripcode (to allow further editing of the post).
// If a hidden tripcode was generated, send a cookie to the client to tell it
// to save the tripcode for further use.
// POST params: text, [nickname]
func apiReply(rw http.ResponseWriter, req *http.Request) {
	// check if this IP is blacklisted
	if isBanned, banReason, blAction := checkBlacklist(req); isBanned {
		switch blAction {
		case "ban":
			sendError(rw, 423, banReason)
			return
		case "captcha":
			// require captcha parameter
			postUserCaptcha := strings.Replace(req.PostFormValue("captcha"), " ", "", -1)
			postUserCaptcha = strings.ToLower(postUserCaptcha)
			postCaptchaQuestion := req.PostFormValue("question")
			found := false
			postRealCaptcha := ""
			for idx := range captchas {
				if captchas[idx].Question == postCaptchaQuestion {
					postRealCaptcha = captchas[idx].Answer
					found = true
					break
				}
			}
			if !found || postUserCaptcha != postRealCaptcha {
				sendError(rw, 401, "Incorrect captcha.")
				return
			}
		}
	}
	vars := mux.Vars(req)
	threadUrl := vars["thread"]
	thread, err := db.GetThread(threadUrl)
	if err != nil {
		fmt.Println(err.Error())
		debug.PrintStack()
		sendError(rw, 500, "Database error: "+err.Error())
		return
	}
	count, err := db.PostCount(&thread)
	if err != nil {
		fmt.Println(err.Error())
		debug.PrintStack()
		sendError(rw, 500, "Database error: "+err.Error())
		return
	}
	page := (count + siteInfo.PostsPerPage - 1) / siteInfo.PostsPerPage
	if page < 1 {
		page = 1
	}

	postNickname := req.PostFormValue("nickname")
	postContent := strings.TrimSpace(req.PostFormValue("text"))
	if len(postContent) < 1 {
		sendError(rw, 400, "Required fields are missing")
		return
	}

	postAntispam := req.PostFormValue("ferrarelle")
	if len(postAntispam) > 0 {
		// Fuck spambots
		sendError(rw, 500, "Database error")
		return
	}

	nickname, tcode := parseNickname(postNickname)
	var hTrip string
	if len(tcode) < 1 {
		hTrip = randomString(8)
		tcode = tripcode(hTrip)
	}
	user := data.User{nickname, tcode, len(hTrip) > 0}
	content := postContent

	if postTooLong(content) {
		sendError(rw, 400, "Post is too long.")
		return
	}
	_, err = db.ReplyThread(&thread, user, content)
	if err != nil {
		fmt.Println(err.Error())
		debug.PrintStack()
		sendError(rw, 500, "Database error: "+err.Error())
		return
	}

	basepath := "/"
	if ok, val := isAdmin(req); ok {
		basepath = val.BasePath
	}

	if len(hTrip) > 0 {
		http.SetCookie(rw, &http.Cookie{
			Name:     "crSetLatestPost",
			Value:    thread.ShortUrl + "/" + strconv.Itoa(count) + "/" + hTrip,
			Path:     "/thread/",
			MaxAge:   600,
			HttpOnly: false,
		})
	}
	http.Redirect(rw, req, basepath+"thread/"+thread.ShortUrl+"/page/"+strconv.Itoa(page)+"#p"+strconv.Itoa(count), http.StatusMovedPermanently)
}

// apiPreview: returns the content that would be inserted in the post if this
// were a reply.
// POST params: text
func apiPreview(rw http.ResponseWriter, req *http.Request) {
	postContent := strings.TrimSpace(req.PostFormValue("text"))
	if len(postContent) < 1 {
		sendError(rw, 400, "Required fields are missing")
		return
	} else if postTooLong(postContent) {
		sendError(rw, 400, "Post is too long")
		return
	}
	content := parseContent(postContent, "bbcode")

	// Do a dummy post for mutators
	var fakepost data.Post
	fakepost.Content = content
	for _, m := range postmutators {
		applyPostMutator(m, nil, &fakepost, req)
	}
	content = fakepost.Content

	fmt.Fprintln(rw, content)
}

// apiEditPost: updates the content of a post and its LastModified field
// (auth via tripcode); if post is OP, also accepts 'tags' param to edit
// thread tags.
// POST params: tripcode, text, [tags]
func apiEditPost(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	isAdmin, val := isAdmin(req)

	postId, err := strconv.Atoi(vars["post"])
	if err != nil {
		sendError(rw, 400, "Invalid post id")
		return
	}

	thread, post, err := threadPostOrErr(rw, vars["thread"], vars["post"])
	if err != nil {
		return
	}
	// Skip a bunch of checks if the user is an admin
	if !isAdmin {
		// if post is deleted and it's not being edited by an admin, refuse
		if post.ContentType == "deleted" || post.ContentType == "admin-deleted" {
			sendError(rw, 403, "Forbidden")
			return
		}
		// if post has no tripcode associated, refuse to edit
		if len(post.Author.Tripcode) < 1 {
			sendError(rw, 403, "Forbidden")
			return
		}
		// check tripcode
		trip := tripcode(req.PostFormValue("tripcode"))
		if trip != post.Author.Tripcode {
			sendError(rw, 401, "Invalid tripcode")
			return
		}
	}
	// update post content and date (strip multiple whitespaces at the end of the text)
	newContent := strings.TrimSpace(req.PostFormValue("text"))
	if len(newContent) < 1 {
		sendError(rw, 400, "Required fields are missing.")
		return
	} else if postTooLong(newContent) {
		sendError(rw, 400, "Post is too long.")
		return
	}
	// update tags if post is OP and 'tags' was passed
	if post.Id == thread.ThreadPost {
		tags := parseTags(req.PostFormValue("tags"))
		if len(tags) > 0 {
			oldtags := make(map[string]bool, len(thread.Tags))
			for _, tag := range thread.Tags {
				oldtags[tag] = true
			}
			err = db.SetThreadTags(thread.Id, tags)
			for _, tag := range tags {
				if oldtags[tag] {
					delete(oldtags, tag)
					continue // no need to inc/dec tag
				}
				// increment new tag
				db.IncTag(tag, thread.Id)
			}
			// decrement any tag which was removed
			for tag := range oldtags {
				db.DecTag(tag)
			}
		}
	}

	err = db.EditPost(post.Id, newContent, "bbcode")
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	basepath := "/"
	if isAdmin {
		basepath = val.BasePath
	}

	page := (postId + siteInfo.PostsPerPage - 1) / siteInfo.PostsPerPage
	if page < 1 {
		page = 1
	}
	http.Redirect(rw, req, basepath+"thread/"+thread.ShortUrl+"/page/"+strconv.Itoa(page)+"#p"+vars["post"], http.StatusMovedPermanently)
}

// apiDeletePost: Sets the 'deleted flag' to a post, auth-ing request by tripcode.
// Original post content is retained in DB (for now)
// POST params: tripcode
func apiDeletePost(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	isAdmin, val := isAdmin(req)

	postId, err := strconv.Atoi(vars["post"])
	if err != nil {
		sendError(rw, 400, "Invalid post id")
		return
	}

	thread, post, err := threadPostOrErr(rw, vars["thread"], vars["post"])
	if err != nil {
		return
	}

	if !isAdmin {
		// if post has no tripcode associated, refuse to delete
		if len(post.Author.Tripcode) < 1 {
			sendError(rw, 403, "Forbidden")
			return
		}
		// check tripcode
		trip := req.PostFormValue("tripcode")
		if tripcode(trip) != post.Author.Tripcode {
			sendError(rw, 401, "Invalid tripcode")
			return
		}
	}

	// are we purging or deleting?
	threadgone := false
	dtype := req.PostFormValue("deletetype")
	if dtype == "purge" {
		if !isAdmin {
			sendError(rw, 403, "Forbidden")
			return
		}
		err, threadgone = db.PurgePost(post)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}
	} else {
		if err = db.DeletePost(post.Id, isAdmin); err != nil {
			sendError(rw, 500, err.Error())
			return
		}
	}

	basepath := "/"
	if isAdmin {
		basepath = val.BasePath
	}

	page := (postId + siteInfo.PostsPerPage - 1) / siteInfo.PostsPerPage
	if page < 1 {
		page = 1
	}

	rurl := basepath
	if !threadgone {
		num, err := strconv.Atoi(vars["post"])
		if err != nil {
			num = 0
		}
		rurl += "thread/" + thread.ShortUrl + "/page/" + strconv.Itoa(page) + "#p" + strconv.Itoa(num-1)
	}
	http.Redirect(rw, req, rurl, http.StatusMovedPermanently)
}

func apiTagSearch(rw http.ResponseWriter, req *http.Request) {
	tags := req.PostFormValue("tags")
	if len(tags) < 1 {
		// if no tags are specified, go back home
		http.Redirect(rw, req, "/", http.StatusNoContent)
		return
	}

	basepath := "/"
	if ok, val := isAdmin(req); ok {
		basepath = val.BasePath
	}

	http.Redirect(rw, req, basepath+"tag/"+url.QueryEscape(tags), http.StatusMovedPermanently)
}

// apiGetRaw: retreive the raw content of a post.
// Refuse to do so if post is deleted and user is not an admin.
// POST params: none
func apiGetRaw(rw http.ResponseWriter, req *http.Request) {
	isAdmin, _ := isAdmin(req)
	vars := mux.Vars(req)
	_, post, err := threadPostOrErr(rw, vars["thread"], vars["post"])
	if err != nil {
		return
	}
	if !isAdmin {
		if post.ContentType == "deleted" || post.ContentType == "admin-deleted" {
			sendError(rw, 403, "Forbidden")
			return
		}
	}
	fmt.Fprintln(rw, post.Content)
}

// apiTagList: get a JSON array containing all tags
// POST params: tag
func apiTagList(rw http.ResponseWriter, req *http.Request) {
	tag := req.PostFormValue("tag")
	_, hTags := getHiddenElems(req)
	tags, err := db.GetMatchingTags(tag, 0, 0, hTags)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}
	tagnames := make([]string, len(tags))
	for i, tag := range tags {
		tagnames[i] = tag.Name
	}
	tagJSON, err := json.Marshal(tagnames)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}
	fmt.Fprintln(rw, string(tagJSON))
}
