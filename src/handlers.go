package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"../mustache"
	. "./data"
	"github.com/gorilla/mux"
)

func httpHome(rw http.ResponseWriter, req *http.Request) {
	hThreads, hTags := getHiddenElems(req)
	tags, err := db.GetPopularTags(siteInfo.HomeTagsNum, 0, hTags)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	tagdata := make([]TagData, len(tags))

	for i := range tags {
		thread, err := db.GetThreadById(tags[i].LastThread)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}
		count, err := db.PostCount(&thread)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}

		name := htmlFullEscape(tags[i].Name)
		tagdata[i] = TagData{
			Name:          name,
			URLName:       url.QueryEscape(name),
			LastUpdate:    tags[i].LastUpdate,
			StrLastUpdate: strdate(tags[i].LastUpdate),
			LastThread: ThreadInfo{
				Thread:      thread,
				LastMessage: count - 1,
				Page:        (count + siteInfo.PostsPerPage - 1) / siteInfo.PostsPerPage,
			},
		}
		if tagdata[i].LastThread.Page < 1 {
			tagdata[i].LastThread.Page = 1
		}
	}

	tinfos, err := retreiveThreads(siteInfo.HomeThreadsNum, 0, hThreads, hTags)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}
	send(rw, req, "home", "", struct {
		Last []ThreadInfo
		Tags []TagData
	}{
		tinfos,
		tagdata,
	})
}

func httpAllThreads(rw http.ResponseWriter, req *http.Request) {
	var pageInt int
	var pageOffset int
	var err error
	vars := mux.Vars(req)

	if page, ok := vars["page"]; ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			sendError(rw, 400, "Invalid page number")
			return
		}
		pageOffset = (pageInt - 1) * siteInfo.ThreadsPerPage
	} else {
		pageInt = 1
		pageOffset = 0
	}

	hThreads, hTags := getHiddenElems(req)
	tinfos, err := retreiveThreads(siteInfo.ThreadsPerPage, pageOffset, hThreads, hTags)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	pages := PageInfo{
		Page:     pageInt,
		HasPrev:  pageInt > 1,
		PrevPage: pageInt - 1,
		HasNext:  len(tinfos) == siteInfo.TagsPerPage,
		NextPage: pageInt + 1,
	}

	send(rw, req, "threads", "All threads", struct {
		Last  []ThreadInfo
		Pages PageInfo
	}{
		tinfos,
		pages,
	})
}

func httpAllTags(rw http.ResponseWriter, req *http.Request) {
	var pageInt int
	var pageOffset int
	var err error
	vars := mux.Vars(req)

	if page, ok := vars["page"]; ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			sendError(rw, 400, "Invalid page number")
			return
		}
		if pageInt < 1 {
			pageInt = 1
		}
		pageOffset = (pageInt - 1) * siteInfo.TagsPerPage
	} else {
		pageInt = 1
		pageOffset = 0
	}

	_, hTags := getHiddenElems(req)
	tags, err := db.GetPopularTags(siteInfo.TagsPerPage, pageOffset, hTags)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}
	tagdata := make([]TagData, len(tags))

	for i := range tags {
		thread, err := db.GetThreadById(tags[i].LastThread)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}
		count, err := db.PostCount(&thread)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}

		name := htmlFullEscape(tags[i].Name)
		tagdata[i] = TagData{
			Name:          name,
			URLName:       url.QueryEscape(name),
			LastUpdate:    tags[i].LastUpdate,
			StrLastUpdate: strdate(tags[i].LastUpdate),
			LastThread: ThreadInfo{
				Thread:      thread,
				LastMessage: count - 1,
				Page:        (count + siteInfo.PostsPerPage - 1) / siteInfo.PostsPerPage,
			},
		}
		if tagdata[i].LastThread.Page < 1 {
			tagdata[i].LastThread.Page = 1
		}
	}

	pages := PageInfo{
		Page:     pageInt,
		HasPrev:  pageInt > 1,
		PrevPage: pageInt - 1,
		HasNext:  len(tags) == siteInfo.TagsPerPage,
		NextPage: pageInt + 1,
	}

	send(rw, req, "tags", "All tags", struct {
		Tags  []TagData
		Pages PageInfo
	}{
		tagdata,
		pages,
	})
}

func httpThread(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	threadUrl := vars["thread"]
	thread, err := db.GetThread(threadUrl)
	isAdmin, _ := isAdmin(req)
	requiresCaptcha := req.Header.Get("Captcha-required") == "true"

	if err != nil {
		if err.Error() == "not found" {
			sendError(rw, 404, nil)
		} else {
			sendError(rw, 500, err.Error())
		}
		return
	}

	var pageInt int
	var pageOffset int
	if page, ok := vars["page"]; ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			sendError(rw, 400, "Invalid page number")
			return
		}
		if pageInt < 1 {
			pageInt = 1
		}
		pageOffset = (pageInt - 1) * siteInfo.PostsPerPage
	} else {
		pageInt = 1
		pageOffset = 0
	}

	postCount, err := db.PostCount(&thread)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	// Escape tags
	for t := range thread.Tags {
		thread.Tags[t] = htmlFullEscape(thread.Tags[t])
	}

	// Parse posts
	posts, err := db.GetPosts(&thread, siteInfo.PostsPerPage, pageOffset)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}
	postsInfo := make([]PostInfo, len(posts))
	for index := range posts {
		posts[index].Content = parseContent(posts[index].Content, posts[index].ContentType)
		// Modules for changing content based on a condition, e.g. Lightify
		for _, m := range postmutators {
			applyPostMutator(m, &thread, &posts[index], &rw, req)
		}
		postsInfo[index].Data = posts[index]
		postsInfo[index].IsDeleted = posts[index].ContentType == "deleted" || posts[index].ContentType == "admin-deleted"
		postsInfo[index].PostId = index + pageOffset
		postsInfo[index].Modified = posts[index].LastModified != 0
		postsInfo[index].Editable = !postsInfo[index].IsDeleted && (isAdmin || !posts[index].Author.HiddenTripcode && len(posts[index].Author.Tripcode) > 0)
		postsInfo[index].StrDate = strdate(posts[index].Date)
		postsInfo[index].StrLastModified = strdate(posts[index].LastModified)
		postsInfo[index].IsAnon = len(posts[index].Author.Nickname) < 1 && (len(posts[index].Author.Tripcode) < 1 || posts[index].Author.HiddenTripcode)
	}

	var threadPost PostInfo
	if pageInt == 1 {
		threadPost = postsInfo[0]
		postsInfo = postsInfo[1:]
	}

	var captchaData CaptchaData
	if requiresCaptcha {
		captchaData, err = randomCaptcha()
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}
	}
	maxPage := (postCount + siteInfo.PostsPerPage - 1) / siteInfo.PostsPerPage
	pages := PageInfo{
		Page:     pageInt,
		HasPrev:  pageInt > 1,
		PrevPage: pageInt - 1,
		HasNext:  pageInt < maxPage,
		NextPage: pageInt + 1,
		MaxPage:  maxPage,
	}

	send(rw, req, "thread", thread.Title, struct {
		Thread       Thread
		ThreadPost   PostInfo
		Posts        []PostInfo
		Pages        PageInfo
		HasOP        bool
		NeedsCaptcha bool
		IsAdmin      bool
		Captcha      CaptchaData
	}{
		thread,
		threadPost,
		postsInfo,
		pages,
		pageInt == 1,
		requiresCaptcha,
		isAdmin,
		captchaData,
	})
}

func httpTagSearch(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	tagName := strings.ToLower(vars["tag"])

	var pageInt int
	var pageOffset int
	var err error
	if page, ok := vars["page"]; ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			sendError(rw, 400, "Invalid page number")
			return
		}
		pageOffset = (pageInt - 1) * siteInfo.TagResultsPerPage
	} else {
		pageInt = 1
		pageOffset = 0
	}

	hThreads, hTags := getHiddenElems(req)
	threads, err := db.GetThreadList(tagName, siteInfo.TagResultsPerPage, pageOffset, hThreads, hTags)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	type ThreadData struct {
		ShortUrl     string
		Title        string
		Author       User
		Tags         []string
		Date         int64
		StrDate      string
		Messages     int32
		ShortContent string
		HasBroken    bool
		LRDate       int64
		LRStrDate    string
		HasLR        bool
		LastPost     struct {
			Author       User
			Date         int64
			StrDate      string
			HasBroken    bool
			ShortContent string
			Number       int
			Page         int
			IsAnon       bool
		}
	}

	threadlist := make([]ThreadData, len(threads))
	for i, v := range threads {
		post, err := db.GetPost(v.ThreadPost)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}

		content := parseContent(post.Content, post.ContentType)
		for _, m := range postmutators {
			applyPostMutator(m, &v, &post, &rw, req)
		}
		short, isbroken := shortify(content)

		// Escape tags
		for t := range v.Tags {
			v.Tags[t] = htmlFullEscape(v.Tags[t])
		}
		threadlist[i] = ThreadData{
			ShortUrl:     v.ShortUrl,
			Title:        v.Title,
			Author:       v.Author,
			Tags:         v.Tags,
			Date:         v.Date,
			StrDate:      strdate(v.Date),
			LRDate:       v.LRDate,
			LRStrDate:    strdate(v.LRDate),
			Messages:     v.Messages - 1,
			ShortContent: short,
			HasBroken:    isbroken,
			HasLR:        v.ThreadPost != v.LastReply,
		}

		if threadlist[i].HasLR {
			reply, err := db.GetPost(v.LastReply)
			if err != nil {
				sendError(rw, 500, err.Error())
				return
			}
			count, err := db.PostCount(&v)
			if err != nil {
				sendError(rw, 500, err.Error())
				return
			}

			lp := &threadlist[i].LastPost
			lp.Author = reply.Author
			lp.Date = reply.Date
			lp.StrDate = strdate(reply.Date)
			content = parseContent(reply.Content, reply.ContentType)
			for _, m := range postmutators {
				applyPostMutator(m, &v, &reply, &rw, req)
			}
			lp.ShortContent, lp.HasBroken = shortify(content)
			lp.Number = count - 1
			lp.Page = (count + siteInfo.PostsPerPage - 2) / siteInfo.PostsPerPage
			lp.IsAnon = len(reply.Author.Nickname) < 1 && (len(reply.Author.Tripcode) < 1 || reply.Author.HiddenTripcode)
		}
	}

	pages := PageInfo{
		Page:     pageInt,
		HasPrev:  pageInt > 1,
		PrevPage: pageInt - 1,
		HasNext:  len(threadlist) == siteInfo.TagResultsPerPage,
		NextPage: pageInt + 1,
	}

	send(rw, req, "tagsearch", "Threads under \""+tagName+"\"", struct {
		ThreadList []ThreadData
		TagName    string
		Pages      PageInfo
	}{
		threadlist,
		url.QueryEscape(tagName),
		pages,
	})
}

func httpNewThread(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Captcha-required") == "true" {
		captchaData, err := randomCaptcha()
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}
		send(rw, req, "newthread", "New thread", struct {
			Captcha CaptchaData
		}{
			captchaData,
		})
	} else {
		send(rw, req, "newthread", "New thread", nil)
	}
}

func httpStiki(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	path := filepath.Base(vars["page"])
	_, err := os.Stat(maudRoot + "/stiki/" + path + ".html")
	if os.IsNotExist(err) {
		sendError(rw, 404, nil)
		return
	} else if err != nil {
		sendError(rw, 500, err.Error())
		return
	}
	stiki(rw, req, path)
}

func httpStikiIndex(rw http.ResponseWriter, req *http.Request) {
	fileList, err := ioutil.ReadDir(maudRoot + "/stiki/")
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	type StikiPage struct {
		PageTitle  string
		PageUrl    string
		LastUpdate int64
		StrDate    string
	}
	var stikiPages []StikiPage
	for _, file := range fileList {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".html") {
			continue
		}
		url := strings.TrimSuffix(file.Name(), ".html")
		// Prettify title
		title := strings.ToUpper(url[:1]) + strings.Replace(url[1:], "-", " ", -1)
		modTime := file.ModTime().Unix()
		if len(title) > 0 {
			stikiPages = append(stikiPages, StikiPage{title, url, modTime, strdate(modTime)})
		}
	}

	send(rw, req, "stiki-index", "Stiki Index", struct {
		Pages []StikiPage
	}{
		stikiPages,
	})
}

func httpManageHidden(rw http.ResponseWriter, req *http.Request) {
	var pageInt int
	var pageOffset int
	var err error
	vars := mux.Vars(req)

	if page, ok := vars["page"]; ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			sendError(rw, 400, "Invalid page number")
			return
		}
		pageOffset = (pageInt - 1) * siteInfo.ThreadsPerPage
	} else {
		pageInt = 1
		pageOffset = 0
	}

	hThreads, hTags := getHiddenElems(req)

	// Get hidden threads
	threads, err := db.GetThreads(hThreads, siteInfo.HomeThreadsNum, pageOffset)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	tinfos := make([]ThreadInfo, len(threads))
	for i := range threads {
		count, err := db.PostCount(&threads[i])
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}

		lastPost, err := db.GetPost(threads[i].LastReply)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}

		tinfos[i] = ThreadInfo{
			Thread:      threads[i],
			LastMessage: count - 1,
			LastPost: PostInfo{
				Data:    lastPost,
				StrDate: strdate(lastPost.Date),
			},
			Page: (count + siteInfo.PostsPerPage - 1) / siteInfo.PostsPerPage,
		}
		if tinfos[i].Page < 1 {
			tinfos[i].Page = 1
		}
	}

	// Get hidden tags
	tags, err := db.GetTags(hTags, siteInfo.HomeTagsNum, pageOffset)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	tagdata := make([]TagData, len(tags))

	for i := range tags {
		thread, err := db.GetThreadById(tags[i].LastThread)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}
		count, err := db.PostCount(&thread)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}

		tagdata[i] = TagData{
			Name:          htmlFullEscape(tags[i].Name),
			LastUpdate:    tags[i].LastUpdate,
			StrLastUpdate: strdate(tags[i].LastUpdate),
			LastThread: ThreadInfo{
				Thread:      thread,
				LastMessage: count - 1,
				Page:        (count + siteInfo.PostsPerPage - 1) / siteInfo.PostsPerPage,
			},
		}
		if tagdata[i].LastThread.Page < 1 {
			tagdata[i].LastThread.Page = 1
		}
	}

	pages := PageInfo{
		Page:     pageInt,
		HasPrev:  pageInt > 1,
		PrevPage: pageInt - 1,
		HasNext:  len(tinfos) == siteInfo.TagResultsPerPage,
		NextPage: pageInt + 1,
	}

	send(rw, req, "hidden", "Hidden elements", struct {
		Last  []ThreadInfo
		Tags  []TagData
		Pages PageInfo
	}{
		tinfos,
		tagdata,
		pages,
	})
}

// httpEditPost serves a page which allows editing posts even when JS is
// disabled on the client. Unless the client is an admin, refuse to edit
// deleted or anon posts.
func httpEditPost(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	isAdmin, _ := isAdmin(req)

	// Retreive post content
	thread, post, err := threadPostOrErr(rw, vars["thread"], vars["post"])
	if err != nil {
		return
	}

	if !isAdmin {
		if post.ContentType == "deleted" || post.ContentType == "admin-deleted" || len(post.Author.Nickname) < 1 {
			sendError(rw, 403, "Forbidden")
			return
		}
	}

	isOp := post.Id == thread.ThreadPost
	var tags string
	if isOp {
		tags = "#" + strings.Join(thread.Tags, " #")
	}

	send(rw, req, "edit", "Edit post", struct {
		Thread   string
		Post     string
		Nickname string
		Content  string
		IsOP     bool
		IsAdmin  bool
		Tags     string
	}{
		vars["thread"],
		vars["post"],
		post.Author.Nickname,
		post.Content,
		isOp,
		isAdmin,
		tags,
	})
}

func httpDeletePost(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	isAdmin, _ := isAdmin(req)

	// Retreive post content
	_, post, err := threadPostOrErr(rw, vars["thread"], vars["post"])
	if err != nil {
		return
	}

	if !isAdmin {
		if post.ContentType == "deleted" || post.ContentType == "admin-deleted" || len(post.Author.Nickname) < 1 {
			sendError(rw, 403, "Forbidden")
			return
		}
	}

	send(rw, req, "delete", "Delete post", struct {
		Thread   string
		Post     string
		Nickname string
		IsAdmin  bool
	}{
		vars["thread"],
		vars["post"],
		post.Author.Nickname,
		isAdmin,
	})
}

func httpBanUser(rw http.ResponseWriter, req *http.Request) {
	isAdmin, _ := isAdmin(req)
	if !isAdmin {
		sendError(rw, 401, "Unauthorized")
		return
	}
	vars := mux.Vars(req)

	_, post, err := threadPostOrErr(rw, vars["thread"], vars["post"])
	if err != nil {
		return
	}

	// retreive Ip
	if len(post.Author.Ip) < 1 {
		sendError(rw, 500, "Couldn't find post IP")
		return
	}

	send(rw, req, "ban", "Ban user", struct {
		Ip string
	}{
		post.Author.Ip,
	})
}

func httpBlacklist(rw http.ResponseWriter, req *http.Request) {
	isAdmin, _ := isAdmin(req)
	if !isAdmin {
		sendError(rw, 401, "Unauthorized")
		return
	}

	// Flattened blacklist structure
	type BlacklistData struct {
		Name      string
		Blacklist Blacklist
	}
	blacklisted := make([]BlacklistData, len(blacklists))
	i := 0
	for name, rule := range blacklists {
		blacklisted[i] = BlacklistData{name, rule}
		i++
	}

	send(rw, req, "blacklist", "Blacklist", struct {
		BLData []BlacklistData
	}{
		blacklisted,
	})
}

func httpRobots(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(rw, "User-agent: *\nDisallow: /thread/*/edit\nDisallow: /thread/*/delete")
}

func httpVars(rw http.ResponseWriter, req *http.Request) {
	vars := map[string]interface{}{
		"domain":   siteInfo.FullVersionDomain,
		"maxlen":   siteInfo.MaxPostLength,
		"basepath": "/",
	}

	if ok, val := isAdmin(req); ok {
		vars["adminMode"] = true
		vars["basepath"] = val.BasePath
	}

	jsonVars, err := json.Marshal(vars)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}
	fmt.Fprintln(rw, "window.crOpts = "+string(jsonVars))
}

func sendCSPHeaders(rw http.ResponseWriter, req *http.Request) {
	head := rw.Header()
	head.Add("Content-Security-Policy", csp.String())
}

func send(rw http.ResponseWriter, req *http.Request, name, title string, context interface{}) {
	sendCSPHeaders(rw, req)
	if len(title) > 0 {
		title = " ~ " + title
	}
	basepath := "/"
	ok, val := isAdmin(req)
	if ok {
		basepath = val.BasePath
	}
	footer := siteInfo.Footer[rand.Intn(len(siteInfo.Footer))]
	if len(siteInfo.PostFooter) > 0 {
		footer += siteInfo.PostFooter
	}
	fmt.Fprintln(rw,
		mustache.RenderFileInLayout(
			maudRoot+"/template/"+name+".html",
			maudRoot+"/template/layout.html",
			struct {
				Info     SiteInfo
				Title    string
				Footer   string
				Data     interface{}
				BasePath string
				UrlPath  string
				IsAdmin  bool
				IsLight  bool
			}{
				siteInfo,
				siteInfo.Title + title,
				footer,
				context,
				basepath,
				req.URL.String(),
				ok,
				isLightVersion(req),
			},
		),
	)
}

func stiki(rw http.ResponseWriter, req *http.Request, name string) {
	sendCSPHeaders(rw, req)
	basepath := "/"
	ok, val := isAdmin(req)
	if ok {
		basepath = val.BasePath
	}
	footer := siteInfo.Footer[rand.Intn(len(siteInfo.Footer))]
	if len(siteInfo.PostFooter) > 0 {
		footer += siteInfo.PostFooter
	}
	title := strings.ToUpper(name[:1]) + strings.Replace(name[1:], "-", " ", -1)
	fmt.Fprintln(rw,
		mustache.RenderFileInLayout(
			maudRoot+"/stiki/"+name+".html",
			maudRoot+"/template/layout.html",
			struct {
				Info     SiteInfo
				Title    string
				Footer   string
				BasePath string
				UrlPath  string
			}{
				siteInfo,
				siteInfo.Title + " ~ Stiki: " + title,
				footer,
				basepath,
				req.URL.String(),
			},
		),
	)
}

func sendError(rw http.ResponseWriter, code int, context interface{}) {
	rw.WriteHeader(code)
	fmt.Fprintln(rw,
		mustache.RenderFileInLayout(
			maudRoot+"/errors/"+strconv.Itoa(code)+".html",
			maudRoot+"/errors/layout.html",
			struct {
				Info SiteInfo
				Data interface{}
			}{
				siteInfo,
				context,
			},
		),
	)
}
