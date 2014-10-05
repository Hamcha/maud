package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
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

func parseContent(content string) string {
	html := blackfriday.MarkdownCommon([]byte(content))
	safe := bluemonday.UGCPolicy().SanitizeBytes(html)
	return string(safe)
}

func parseTags(tags string) []string {
	if len(tags) < 1 {
		return nil
	}
	list := strings.Split(tags, ",")
	for i := range list {
		// Spaces begone
		list[i] = strings.TrimSpace(list[i])
		// Strip initial # if any
		if list[i][0] == '#' {
			list[i] = list[i][1:]
		}
	}
	return list
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
