package main

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"strings"
)

func parseNickname(nickname string) (string, string) {
	if len(nickname) < 1 {
		return "", ""
	}
	nicks := strings.SplitN(nickname, "#", 2)
	if len(nicks) > 1 {
		return nicks[0], nicks[1]
	}
	return nickname, ""
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
		list[i] = strings.TrimSpace(list[i])
	}
	return list
}

func generateURL(timestamp int64) string {
	buf := make([]byte, 8)
	binary.PutVarint(buf, timestamp)
	str := hex.EncodeToString(buf)
	return strings.TrimRight(str, "0")
}
