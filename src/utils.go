package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net/http"
	"regexp"
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
		safe := PostPolicy().Sanitize(content)
		bbc := bbcode(safe)
		html := ParseMarkdown(bbc)
		return html
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
	list := strings.Split(tags, ",")
	for i := range list {
		// Spaces begone
		list[i] = strings.ToLower(strings.TrimSpace(list[i]))
		// Strip initial # if any
		if list[i][0] == '#' {
			list[i] = list[i][1:]
		}
	}
	list = removeDuplicates(list)
	return list
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

func generateURL(name string) string {
	buf := make([]byte, 8)
	num, _ := DBNextId(name)
	binary.PutVarint(buf, num+1)
	btr := bytes.TrimRight(buf, "\000")
	str := base64.URLEncoding.EncodeToString(btr)
	return strings.TrimRight(str, "=")
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
	thread, err := DBGetThread(threadId)
	// retreive post
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		http.Error(rw, "Invalid post ID", 400)
		return thread, Post{}, err
	}
	posts, err := DBGetPosts(&thread, 1, postId)
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

func lightify(content string) string {
	img := regexp.MustCompile("(?:<a [^>]+>)?<img .*src=(\"[^\"]+\"|'[^']+'|[^'\"][^\\s]+).*>(?:</a>)?")
	content = img.ReplaceAllString(content, "<a class='toggleImage' data-url=$1>[Click to view image]</a>")
	iframe := regexp.MustCompile("<iframe .*src=(\"[^\"]+\"|'[^']+'|[^'\"][^\\s]+).*>")
	content = iframe.ReplaceAllString(content, "<a target=\"_blank\" href=$1>[Click to open embedded content]</a>")
	return content
}
