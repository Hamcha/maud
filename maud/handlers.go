package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/viper"

	"github.com/gorilla/mux"
	. "github.com/hamcha/maud/maud/data"
	"github.com/hamcha/maud/mustache"
)

func httpHome(rw http.ResponseWriter, req *http.Request) {
	hThreads, hTags := getHiddenElems(req)
	tags, err := db.GetPopularTags(viper.GetInt("tagsInHome"), 0, hTags)
	if err != nil {
		send500(rw, err)
		return
	}

	tagdata := make([]TagData, 0)

	for i := range tags {
		thread, err := db.GetThreadById(tags[i].LastThread)
		var deleted bool
		for err != nil {
			// Oh fuck, try to fix or hide the tag
			deleted, err = db.HealTag(tags[i].Name)
			if err != nil {
				send500(rw, err)
				return
			}
			if deleted {
				break
			} else {
				// Re-fetch tag
				tag, err := db.GetTag(tags[i].Name)
				if err != nil {
					// WTF
					deleted = true
					break
				}
				thread, err = db.GetThreadById(tag.LastThread)
			}
		}
		if deleted {
			continue
		}
		count, err := db.PostCount(&thread)
		if err != nil {
			// Should probably check why this is broken
			count = 0
			continue
		}

		name := htmlFullEscape(tags[i].Name)
		postsPerPage := viper.GetInt("postsPerPage")
		tagdata = append(tagdata, TagData{
			Name:          name,
			URLName:       url.QueryEscape(name),
			LastUpdate:    tags[i].LastUpdate,
			StrLastUpdate: strdate(tags[i].LastUpdate),
			LastThread: ThreadInfo{
				Thread:      thread,
				LastMessage: count - 1,
				Page:        (count + postsPerPage - 1) / postsPerPage,
			},
		})
		if tagdata[len(tagdata)-1].LastThread.Page < 1 {
			tagdata[len(tagdata)-1].LastThread.Page = 1
		}
	}

	tinfos, err := retreiveThreads(viper.GetInt("threadsInHome"), 0, hThreads, hTags)
	if err != nil {
		send500(rw, err)
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
	threadsPerPage := viper.GetInt("threadsPerPage")

	if page, ok := vars["page"]; ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			sendError(rw, 400, "Invalid page number")
			return
		}
		pageOffset = (pageInt - 1) * threadsPerPage
	} else {
		pageInt = 1
		pageOffset = 0
	}

	hThreads, hTags := getHiddenElems(req)
	tinfos, err := retreiveThreads(threadsPerPage, pageOffset, hThreads, hTags)
	if err != nil {
		send500(rw, err)
		return
	}

	pages := PageInfo{
		Page:     pageInt,
		HasPrev:  pageInt > 1,
		PrevPage: pageInt - 1,
		HasNext:  len(tinfos) == viper.GetInt("tagsPerPage"),
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
	tagsPerPage := viper.GetInt("tagsPerPage")
	postsPerPage := viper.GetInt("postsPerPage")

	if page, ok := vars["page"]; ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			sendError(rw, 400, "Invalid page number")
			return
		}
		if pageInt < 1 {
			pageInt = 1
		}
		pageOffset = (pageInt - 1) * tagsPerPage
	} else {
		pageInt = 1
		pageOffset = 0
	}

	_, hTags := getHiddenElems(req)
	tags, err := db.GetPopularTags(tagsPerPage, pageOffset, hTags)
	if err != nil {
		send500(rw, err)
		return
	}
	tagdata := make([]TagData, 0)

	for i := range tags {
		thread, err := db.GetThreadById(tags[i].LastThread)
		var deleted bool
		for err != nil {
			// Oh fuck, try to fix or hide the tag
			deleted, err = db.HealTag(tags[i].Name)
			if err != nil {
				send500(rw, err)
				return
			}
			if deleted {
				break
			} else {
				// Re-fetch tag
				tag, err := db.GetTag(tags[i].Name)
				if err != nil {
					// WTF
					deleted = true
					break
				}
				thread, err = db.GetThreadById(tag.LastThread)
			}
		}
		if deleted {
			continue
		}
		count, err := db.PostCount(&thread)
		if err != nil {
			// Should probably check why this is broken
			count = 0
			return
		}

		name := htmlFullEscape(tags[i].Name)
		tagdata = append(tagdata, TagData{
			Name:          name,
			URLName:       url.QueryEscape(name),
			LastUpdate:    tags[i].LastUpdate,
			StrLastUpdate: strdate(tags[i].LastUpdate),
			LastThread: ThreadInfo{
				Thread:      thread,
				LastMessage: count - 1,
				Page:        (count + postsPerPage - 1) / postsPerPage,
			},
		})
		if tagdata[len(tagdata)-1].LastThread.Page < 1 {
			tagdata[len(tagdata)-1].LastThread.Page = 1
		}
	}

	pages := PageInfo{
		Page:     pageInt,
		HasPrev:  pageInt > 1,
		PrevPage: pageInt - 1,
		HasNext:  len(tags) == tagsPerPage,
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
	postsPerPage := viper.GetInt("postsPerPage")

	if err != nil {
		if err.Error() == "not found" {
			sendError(rw, 404, nil)
		} else {
			send500(rw, err)
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
		pageOffset = (pageInt - 1) * postsPerPage
	} else {
		pageInt = 1
		pageOffset = 0
	}

	postCount, err := db.PostCount(&thread)
	if err != nil {
		send500(rw, err)
		return
	}

	// Escape tags
	for t := range thread.Tags {
		thread.Tags[t] = htmlFullEscape(thread.Tags[t])
	}

	// Parse posts
	posts, err := db.GetPosts(&thread, postsPerPage, pageOffset)
	if err != nil {
		send500(rw, err)
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
		postsInfo[index].Modified = posts[index].LastModified != 0
		postsInfo[index].Editable = !postsInfo[index].IsDeleted &&
			(isAdmin || !posts[index].Author.HiddenTripcode && len(posts[index].Author.Tripcode) > 0)
		postsInfo[index].StrDate = strdate(posts[index].Date)
		postsInfo[index].SchemaDate = time.Unix(posts[index].Date, 0).Format(time.RFC3339)
		postsInfo[index].StrLastModified = strdate(posts[index].LastModified)
		postsInfo[index].SchemaLastModified = time.Unix(posts[index].LastModified, 0).Format(time.RFC3339)
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
			send500(rw, err)
			return
		}
	}
	maxPage := (postCount + postsPerPage - 1) / postsPerPage
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
		EmojiLink    string
	}{
		thread,
		threadPost,
		postsInfo,
		pages,
		pageInt == 1,
		requiresCaptcha,
		isAdmin,
		captchaData,
		emojiLink(threadUrl),
	})
}

func httpTagSearch(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	tagName := strings.ToLower(vars["tag"])
	tagResultsPerPage := viper.GetInt("tagResultsPerPage")
	postsPerPage := viper.GetInt("postsPerPage")

	var pageInt int
	var pageOffset int
	var err error
	if page, ok := vars["page"]; ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			sendError(rw, 400, "Invalid page number")
			return
		}
		pageOffset = (pageInt - 1) * tagResultsPerPage
	} else {
		pageInt = 1
		pageOffset = 0
	}

	hThreads, hTags := getHiddenElems(req)

	tagReadable, err := url.QueryUnescape(tagName)
	if err != nil {
		send500(rw, err)
		return
	}
	tagReadable = html.UnescapeString(tagReadable)
	tagReadable = strings.TrimSuffix(tagReadable, " #")

	threads, err := db.GetThreadList(tagReadable, tagResultsPerPage, pageOffset, hThreads, hTags)
	if err != nil {
		send500(rw, err)
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
			send500(rw, err)
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
				send500(rw, err)
				return
			}
			count, err := db.PostCount(&v)
			if err != nil {
				send500(rw, err)
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
			lp.Page = (count + postsPerPage - 2) / postsPerPage
			lp.IsAnon = len(reply.Author.Nickname) < 1 && (len(reply.Author.Tripcode) < 1 || reply.Author.HiddenTripcode)
		}
	}

	pages := PageInfo{
		Page:     pageInt,
		HasPrev:  pageInt > 1,
		PrevPage: pageInt - 1,
		HasNext:  len(threadlist) == tagResultsPerPage,
		NextPage: pageInt + 1,
	}

	send(rw, req, "tagsearch", "Threads under \""+tagName+"\"", struct {
		ThreadList      []ThreadData
		TagNameQuery    string
		TagNameReadable string
		Pages           PageInfo
	}{
		threadlist,
		url.QueryEscape(tagName),
		tagReadable,
		pages,
	})
}

func httpNewThread(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Captcha-required") == "true" {
		captchaData, err := randomCaptcha()
		if err != nil {
			send500(rw, err)
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
	fpath := filepath.Base(vars["page"])
	_, err := os.Stat(path.Join(maudRoot, "stiki", fpath+".html"))
	if os.IsNotExist(err) {
		sendError(rw, 404, nil)
		return
	} else if err != nil {
		send500(rw, err)
		return
	}
	stiki(rw, req, fpath)
}

func httpStikiIndex(rw http.ResponseWriter, req *http.Request) {
	fileList, err := ioutil.ReadDir(path.Join(maudRoot, "stiki"))
	if err != nil {
		send500(rw, err)
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
	threadsPerPage := viper.GetInt("threadsPerPage")
	postsPerPage := viper.GetInt("postsPerPage")
	tagResultsPerPage := viper.GetInt("tagResultsPerPage")
	threadsInHome := viper.GetInt("threadsInHome")
	tagsInHome := viper.GetInt("tagsInHome")

	if page, ok := vars["page"]; ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			sendError(rw, 400, "Invalid page number")
			return
		}
		pageOffset = (pageInt - 1) * threadsPerPage
	} else {
		pageInt = 1
		pageOffset = 0
	}

	hThreads, hTags := getHiddenElems(req)

	// Get hidden threads
	threads, err := db.GetThreads(hThreads, threadsInHome, pageOffset)
	if err != nil {
		send500(rw, err)
		return
	}

	tinfos := make([]ThreadInfo, len(threads))
	for i := range threads {
		count, err := db.PostCount(&threads[i])
		if err != nil {
			send500(rw, err)
			return
		}

		lastPost, err := db.GetPost(threads[i].LastReply)
		if err != nil {
			send500(rw, err)
			return
		}

		tinfos[i] = ThreadInfo{
			Thread:      threads[i],
			LastMessage: count - 1,
			LastPost: PostInfo{
				Data:    lastPost,
				StrDate: strdate(lastPost.Date),
			},
			Page: (count + postsPerPage - 1) / postsPerPage,
		}
		if tinfos[i].Page < 1 {
			tinfos[i].Page = 1
		}
	}

	// Get hidden tags
	tags, err := db.GetTags(hTags, tagsInHome, pageOffset)
	if err != nil {
		send500(rw, err)
		return
	}

	tagdata := make([]TagData, len(tags))

	for i := range tags {
		thread, err := db.GetThreadById(tags[i].LastThread)
		if err != nil {
			send500(rw, err)
			return
		}
		count, err := db.PostCount(&thread)
		if err != nil {
			send500(rw, err)
			return
		}

		tagdata[i] = TagData{
			Name:          htmlFullEscape(tags[i].Name),
			LastUpdate:    tags[i].LastUpdate,
			StrLastUpdate: strdate(tags[i].LastUpdate),
			LastThread: ThreadInfo{
				Thread:      thread,
				LastMessage: count - 1,
				Page:        (count + postsPerPage - 1) / postsPerPage,
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
		HasNext:  len(tinfos) == tagResultsPerPage,
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
		send500(rw, errors.New("Couldn't find post IP"))
		return
	}

	send(rw, req, "ban", "Ban user", struct {
		Ip string
	}{
		"^" + strings.Replace(post.Author.Ip, `.`, `\.`, -1),
	})
}

type BlacklistDataList []BlacklistData

// Flattened blacklist structure
type BlacklistData struct {
	Name      string
	Blacklist Blacklist
}

func (blData BlacklistDataList) Len() int {
	return len(blData)
}

func (blData BlacklistDataList) Less(i, j int) bool {
	return blData[i].Name < blData[j].Name
}

func (blData BlacklistDataList) Swap(i, j int) {
	blData[i], blData[j] = blData[j], blData[i]
}

func httpBlacklist(rw http.ResponseWriter, req *http.Request) {
	isAdmin, _ := isAdmin(req)
	if !isAdmin {
		sendError(rw, 401, "Unauthorized")
		return
	}

	blacklisted := make(BlacklistDataList, len(blacklists))
	i := 0
	for name, rule := range blacklists {
		blacklisted[i] = BlacklistData{name, rule}
		i++
	}

	sort.Sort(blacklisted)

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
		"domain":   viper.GetString("fullDomain"),
		"maxlen":   viper.GetString("maxPostLength"),
		"basepath": "/",
	}

	if ok, val := isAdmin(req); ok {
		vars["adminMode"] = true
		vars["basepath"] = val.BasePath
	}

	jsonVars, err := json.Marshal(vars)
	if err != nil {
		send500(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(rw, "window.crOpts = %s;\nObject.freeze(window.crOpts);\n", string(jsonVars))
}

func sendCSPHeaders(rw http.ResponseWriter, req *http.Request) {
	head := rw.Header()
	for name, vals := range csp {
		head.Add("Content-Security-Policy", name+" "+vals)
	}
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
	footer := footers[rand.Intn(len(footers))]
	if viper.IsSet("postFooter") {
		footer += viper.GetString("postFooter")
	}
	fmt.Fprintln(rw,
		mustache.RenderFileInLayout(
			path.Join(maudRoot, "template", name+".html"),
			path.Join(maudRoot, "template", "layout.html"),
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
				getSiteInfo(),
				viper.GetString("siteTitle") + title,
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
	footer := footers[rand.Intn(len(footers))]
	if viper.IsSet("postFooter") {
		footer += viper.GetString("postFooter")
	}
	title := strings.ToUpper(name[:1]) + strings.Replace(name[1:], "-", " ", -1)
	fmt.Fprintln(rw,
		mustache.RenderFileInLayout(
			path.Join(maudRoot, "stiki", name+".html"),
			path.Join(maudRoot, "template", "layout.html"),
			struct {
				Info     SiteInfo
				Title    string
				Footer   string
				BasePath string
				UrlPath  string
			}{
				getSiteInfo(),
				viper.GetString("siteTitle") + " ~ Stiki: " + title,
				footer,
				basepath,
				req.URL.String(),
			},
		),
	)
}

func send500(rw http.ResponseWriter, err error) {
	log.Printf("Encountered server error: %s\n", err.Error())
	debug.PrintStack()
	errdata := fmt.Sprintf("Error message: %s\nStack: %s", err.Error(), debug.Stack())
	errdatab64 := base64.StdEncoding.EncodeToString([]byte(errdata))
	sendError(rw, 500, errdatab64)
	// Server errors are pretty bad
	panic(err)
}

func sendError(rw http.ResponseWriter, code int, context interface{}) {
	rw.WriteHeader(code)
	fmt.Fprintln(rw,
		mustache.RenderFileInLayout(
			path.Join(maudRoot, "errors", strconv.Itoa(code)+".html"),
			path.Join(maudRoot, "errors", "layout.html"),
			struct {
				Info SiteInfo
				Data interface{}
			}{
				getSiteInfo(),
				context,
			},
		),
	)
}

// Emoji short url checker and redirector

func emojiRedir(req *http.Request) (bool, string) {
	path := []byte(req.URL.Path[1:])
	var data []byte
	for len(path) > 0 {
		r, size := utf8.DecodeRune(path)
		pos := -1
		for i, emoji := range emojis {
			if emoji == r {
				pos = i
				break
			}
		}
		if pos < 0 {
			return false, ""
		}

		data = append(data, byte(pos))
		path = path[size:]
	}
	num, _ := binary.Varint(data)

	return true, toB64(num)
}
