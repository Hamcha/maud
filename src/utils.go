package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func parseNickname(nickname string) (string, string) {
	if len(nickname) < 1 {
		return "", ""
	}
	nicks := strings.SplitN(nickname, "#", 2)
	if len(nicks) > 1 {
		return nicks[0], tripcode(nicks[1])
	}
	return nickname, ""
}

func tripcode(str string) string {
	sum := sha256.Sum256([]byte(str + siteInfo.Secret))
	b64 := base64.URLEncoding.EncodeToString(sum[:])
	return b64[0:6]
}

func parseContent(content, ctype string) string {
	switch ctype {
	/* New and hot BBcode + Markdown */
	case "bbcode":
		code := PostPolicy().Sanitize(content)
		for _, f := range formatters {
			code = f.Format(code)
		}
		return code
	/* Deleted posts */
	case "deleted":
		return "<em>Message deleted by the user</em>"
	case "admin-deleted":
		return "<em>Message deleted by an admin</em>"
	/* Old and busted preparsed */
	default:
		return content
	}
}

func parseTags(tags string) []string {
	if len(tags) < 1 {
		return nil
	}
	list := strings.Split(tags, "#")
	list = removeEmpty(list)
	for i := range list {
		// Spaces begone
		list[i] = strings.ToLower(strings.TrimSpace(list[i]))
	}
	list = removeDuplicates(list)
	return list
}

func removeEmpty(in []string) []string {
	out := make([]string, 0)
	for _, i := range in {
		if len(strings.TrimSpace(i)) > 0 {
			out = append(out, i)
		}
	}

	return out
}

func removeDuplicates(in []string) []string {
	out := make([]string, 0)
	for _, i := range in {
		found := false
		for _, j := range out {
			if i == j {
				found = true
				break
			}
		}

		if !found {
			out = append(out, i)
		}
	}

	return out
}

func shortify(content string) (string, bool) {
	if len(content) < 300 {
		return content, false
	}

	// count open HTML tags in content
	short := content[:300]
	stack := make([]string, 0)
	stackindex := -1
	offset := -1
	for offset < len(short) {
		offset = index(short, offset+1, '<')
		if offset < 0 {
			break
		}
		end := index(short, offset+1, '>')
		if end < 0 {
			break
		}
		tagname := short[offset+1 : end]

		if tagname[0] == '/' {
			if stackindex < 1 {
				continue
			}
			if tagname[1:] == stack[stackindex] {
				stack = stack[:stackindex]
				stackindex--
			}
		} else {
			stack = append(stack, tagname)
			stackindex++
		}
	}
	// close unclosed tags
	for stackindex > 0 {
		short += "</" + stack[stackindex] + ">"
		stackindex--
	}

	return PostPolicy().Sanitize(short), true
}

func threadPostOrErr(rw http.ResponseWriter, threadId, postIdStr string) (Thread, Post, error) {
	thread, err := db.GetThread(threadId)
	// retreive post
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		sendError(rw, 400, err.Error())
		return thread, Post{}, err
	}
	posts, err := db.GetPosts(&thread, 1, postId)
	if err != nil {
		sendError(rw, 500, err.Error())
		return thread, posts[0], err
	}
	if len(posts) < 1 {
		sendError(rw, 404, "Post not found")
		return thread, posts[0], errors.New("Post not found")
	}
	return thread, posts[0], nil
}

func postTooLong(content string) bool {
	return siteInfo.MaxPostLength > 0 && utf8.RuneCountInString(content) > siteInfo.MaxPostLength
}

func filterFromCookie(req *http.Request) []string {
	cookie, err := req.Cookie("filter")
	if err != nil {
		return nil
	}
	return strings.Split(cookie.String(), ":")
}

func isLightVersion(req *http.Request) bool {
	return len(siteInfo.LightVersionDomain) > 0 && req.Host == siteInfo.LightVersionDomain
}

func index(str string, offset int, del uint8) int {
	for i := offset; i < len(str); i++ {
		if str[i] == del {
			return i
		}
	}
	return -1
}

func generateURL(db Database, name string) string {
	buf := make([]byte, 8)
	num, _ := db.NextId(name)
	binary.PutVarint(buf, num+1)
	btr := bytes.TrimRight(buf, "\000")
	str := base64.URLEncoding.EncodeToString(btr)
	return strings.TrimRight(str, "=")
}

func randomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to a quite less secure method
		sum := sha256.Sum256([]byte(strconv.Itoa(int(time.Now().UnixNano())) + siteInfo.Secret))
		b64 := base64.URLEncoding.EncodeToString(sum[:])
		for len(b64) < length {
			sum2 := sha256.Sum256(sum[:])
			b64 = base64.URLEncoding.EncodeToString(sum2[:]) + b64
		}
		return b64[0:length]
	}
	return base64.URLEncoding.EncodeToString(b)[0:length]
}

func strdate(unixtime int64) string {
	return time.Unix(unixtime, 0).Format("02/01/2006 15:04")
}

// getHiddenElems checks if the request contains a cookie 'crHidden'
// and parses its value, returning a slice with shorturls of hidden threads.
// If no cookie exists, or its value is invalid, err != nil is returned.
func getHiddenElems(req *http.Request) ([]string, error) {
	if cookie, err := req.Cookie("crHidden"); err == nil {
		// cookie value has the format: "url1 url2 url3 ..."
		val, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			return nil, err
		}
		threads := strings.Split(val, " ")
		return threads, err
	} else {
		return nil, err
	}
}

func retreiveThreads(n, offset int, req *http.Request, filter []string) ([]ThreadInfo, error) {
	// check if some threads have been hidden by the user
	hiddenThreads, err := getHiddenElems(req)

	extra := 0
	if err == nil {
		extra = len(hiddenThreads)
	}

	threads, err := db.GetThreadList("", n+extra, offset, filter)
	if err != nil {
		return nil, err
	}

	tinfos := make([]ThreadInfo, 0, siteInfo.HomeThreadsNum)
Outer:
	for i, _ := range threads {
		for _, hidden := range hiddenThreads {
			if threads[i].ShortUrl == hidden {
				continue Outer
			}
		}

		count, err := db.PostCount(&threads[i])
		if err != nil {
			return tinfos, err
		}

		lastPost, err := db.GetPost(threads[i].LastReply)
		if err != nil {
			return tinfos, err
		}

		tinfos = append(tinfos, ThreadInfo{
			Thread:      threads[i],
			LastMessage: count - 1,
			LastPost: PostInfo{
				Data:    lastPost,
				StrDate: strdate(lastPost.Date),
			},
			Page: (count + siteInfo.PostsPerPage - 1) / siteInfo.PostsPerPage,
		})
		j := len(tinfos) - 1
		if tinfos[j].Page < 1 {
			tinfos[j].Page = 1
		}
		if j == n-1 {
			break
		}
	}
	return tinfos, err
}
