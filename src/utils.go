package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net/http"
	"strconv"
	"strings"
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
		if len(i) > 0 {
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
		http.Error(rw, "Invalid post ID", 400)
		return thread, Post{}, err
	}
	posts, err := db.GetPosts(&thread, 1, postId)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return thread, posts[0], err
	}
	if len(posts) < 1 {
		http.Error(rw, "Post not found", 404)
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

func sanitizeURL(tags string) string {
	// replace all characters which are dangerous in an URL
	sane := strings.Replace(tags, "/", "&sol;", -1)
	sane = strings.Replace(sane, "#", "&num;", -1)
	sane = strings.Replace(sane, "?", "&quest;", -1)
	return sane
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
