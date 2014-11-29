package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"math/rand"
	"strings"
	"time"
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

func seedRand() {
	rand.Seed(time.Now().UnixNano())
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

	return PostPolicy().Sanitize(content[:300]), true
}
