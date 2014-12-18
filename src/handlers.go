package main

import (
	"../mustache"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

func httpHome(rw http.ResponseWriter, req *http.Request) {
	tags, err := DBGetPopularTags(10)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
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

	threads, err := DBGetThreadList("", 5, 0)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}

	send(rw, req, "home", "", struct {
		Last []Thread
		Tags []TagData
	}{
		threads,
		tagdata,
	})
}

func httpAllThreads(rw http.ResponseWriter, req *http.Request) {
	threads, err := DBGetThreadList("", 0, 0)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}
	send(rw, req, "threads", "All threads", struct {
		Last []Thread
	}{
		threads,
	})
}

func httpAllTags(rw http.ResponseWriter, req *http.Request) {
	tags, err := DBGetPopularTags(0)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
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

	send(rw, req, "tags", "All tags", struct {
		Tags []TagData
	}{
		tagdata,
	})
}

func httpThread(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	threadUrl := vars["thread"]
	thread, err := DBGetThread(threadUrl)
	isAdmin, _ := isAdmin(req)

	if err != nil {
		if err.Error() == "not found" {
			sendError(rw, 404, nil)
		} else {
			sendError(rw, 500, err.Error())
		}
		return
	}

	// Parse posts
	type PostInfo struct {
		PostId    int
		Data      Post
		IsDeleted bool
		Editable  bool
	}
	posts, err := DBGetPosts(&thread, 0, 0)
	if err != nil {
		sendError(rw, 500, err.Error())
		return
	}
	postsInfo := make([]PostInfo, len(posts))
	for index := range posts {
		posts[index].Content = parseContent(posts[index].Content, posts[index].ContentType)
		postsInfo[index].Data = posts[index]
		postsInfo[index].IsDeleted = posts[index].ContentType == "deleted" || posts[index].ContentType == "admin-deleted"
		postsInfo[index].PostId = index
		postsInfo[index].Editable = !postsInfo[index].IsDeleted && (isAdmin || len(posts[index].Author.Tripcode) > 0)
	}

	send(rw, req, "thread", thread.Title, struct {
		Thread     Thread
		ThreadPost PostInfo
		Posts      []PostInfo
	}{
		thread,
		postsInfo[0],
		postsInfo[1:],
	})
}

func httpTagSearch(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	tagName := strings.ToLower(vars["tag"])

	threads, err := DBGetThreadList(tagName, 0, 0)
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
		Messages     int32
		ShortContent string
		HasBroken    bool
		LRDate       int64
		HasLR        bool
		LastPost     struct {
			Author       User
			Date         int64
			HasBroken    bool
			ShortContent string
		}
	}

	threadlist := make([]ThreadData, len(threads))
	for i, v := range threads {
		post, err := DBGetPost(v.ThreadPost)
		if err != nil {
			sendError(rw, 500, err.Error())
			return
		}

		short, isbroken := shortify(parseContent(post.Content, post.ContentType))

		threadlist[i] = ThreadData{
			ShortUrl:     v.ShortUrl,
			Title:        v.Title,
			Author:       v.Author,
			Tags:         v.Tags,
			Date:         v.Date,
			LRDate:       v.LRDate,
			Messages:     v.Messages - 1,
			ShortContent: short,
			HasBroken:    isbroken,
			HasLR:        v.ThreadPost != v.LastReply,
		}

		if threadlist[i].HasLR {
			reply, err := DBGetPost(v.LastReply)
			if err != nil {
				sendError(rw, 500, err.Error())
				return
			}

			lp := &threadlist[i].LastPost
			lp.Author = reply.Author
			lp.Date = reply.Date
			lp.ShortContent, lp.HasBroken = shortify(parseContent(reply.Content, reply.ContentType))
		}
	}

	send(rw, req, "tagsearch", "Threads under \""+tagName+"\"", threadlist)
}

func httpNewThread(rw http.ResponseWriter, req *http.Request) {
	send(rw, req, "newthread", "New thread", nil)
}

func send(rw http.ResponseWriter, req *http.Request, name string, title string, context interface{}) {
	if len(title) > 0 {
		title = " ~ " + title
	}
	basepath := "/"
	ok, val := isAdmin(req)
	if ok {
		basepath = val.BasePath
	}
	fmt.Fprintln(rw,
		mustache.RenderFileInLayout(
			maudRoot+"/template/"+name+".html",
			maudRoot+"/template/layout.html",
			struct {
				Info          SiteInfo
				Title         string
				MaxPostLength int
				Data          interface{}
				BasePath      string
				IsAdmin       bool
			}{
				siteInfo,
				siteInfo.Title + title,
				siteInfo.MaxPostLength,
				context,
				basepath,
				ok,
			}))
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
			}))
}
