package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"html"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	. "github.com/hamcha/maud/maud/data"
)

var (
	errPostNotFound         = errors.New("Post not found")
	errCapthasNotConfigured = errors.New("Sorry, captchas weren't configured properly.")
)

func parseNickname(nickname string) (string, string) {
	nickname = strings.TrimSpace(nickname)
	if len(nickname) < 1 {
		return "", ""
	}
	nicks := strings.SplitN(nickname, "#", 2)
	if len(nicks) > 1 && len(nicks[0]) > 0 && len(nicks[1]) > 0 {
		return nicks[0], tripcode(nicks[1])
	}
	return nickname, ""
}

func tripcode(str string) string {
	sum := sha256.Sum256([]byte(str + siteInfo.Secret))
	b64 := base64.URLEncoding.EncodeToString(sum[:6])
	return b64[:6]
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
	// Strip disallowed characters
	tags = strings.Replace(tags, "&", "", -1)
	if len(tags) < 1 {
		return nil
	}
	list := strings.Split(tags, "#")
	list = removeEmpty(list)
	for i := range list {
		// Spaces begone
		list[i] = strings.ToLower(strings.TrimSpace(list[i]))
		// limit tag length
		if len(list[i]) > 31 {
			list[i] = list[i][:31]
		}
	}
	list = removeDuplicates(list)
	return list
}

func removeEmpty(in []string) []string {
	var out []string
	for _, i := range in {
		if len(strings.TrimSpace(i)) > 0 {
			out = append(out, i)
		}
	}

	return out
}

func removeDuplicates(in []string) []string {
	var out []string
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

// shortify returns a string which is either `content` or its first
// ~300 runes, ensuring all HTML tags are properly closed and the
// returned string is sanitized. The second return parameter is
// false if the content wasn't shortified, true otherwise.
func shortify(content string) (string, bool) {
	if len(content) < 300 {
		return content, false
	}

	// count open HTML tags in content
	short := content[:300]
	var stack []string
	stackindex := -1
	offset := -1
	// Saves the offset before the latest tag opening
	preTagOffset := -1
	isTagOpen := false
	for offset < len(short) {
		offset = index(short, offset+1, '<')
		preTagOffset = offset - 1
		if offset < 0 {
			break
		}
		isTagOpen = true
		end := index(short, offset+1, '>')
		if end < 0 {
			break
		}
		isTagOpen = false
		tagname := short[offset+1 : end]

		if tagname[0] == '/' {
			// it's a closing tag
			if stackindex < 0 {
				continue
			}
			if tagname[1:] == stack[stackindex] {
				stack = stack[:stackindex]
				stackindex--
			}
		} else if tagname[len(tagname)-1] != '/' {
			// don't take self-closing tags into account
			spaceidx := strings.IndexFunc(tagname, unicode.IsSpace)
			if spaceidx > 0 {
				tagname = tagname[:spaceidx]
			}
			stack = append(stack, tagname)
			stackindex++
		}
	}

	if isTagOpen && preTagOffset > 0 {
		// we broke from the cycle while inside a tag: drop it.
		short = short[:preTagOffset]
	}

	// close unclosed tags
	for stackindex >= 0 {
		short += "</" + stack[stackindex] + ">"
		stackindex--
	}

	return PostPolicy().Sanitize(short), true
}

func threadPostOrErr(rw http.ResponseWriter, threadId, postIdStr string) (Thread, Post, error) {
	thread, err := db.GetThread(threadId)
	if err != nil {
		sendError(rw, 404, "Thread not found")
		return Thread{}, Post{}, err
	}

	// parse post ID
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		sendError(rw, 400, err.Error())
		return thread, Post{}, err
	}

	// special case: OP
	if postId == 0 {
		post, err := db.GetPost(thread.ThreadPost)
		if err != nil {
			sendError(rw, 500, err.Error())
			return thread, Post{}, err
		}
		return thread, post, err
	}

	posts, err := db.GetPosts(&thread, 1, postId)
	if err != nil {
		sendError(rw, 500, err.Error())
		return thread, posts[0], err
	}
	if len(posts) < 1 {
		sendError(rw, 404, "Post not found")
		return thread, Post{}, errPostNotFound
	}
	return thread, posts[0], nil
}

func postTooLong(content string) bool {
	return siteInfo.MaxPostLength > 0 && utf8.RuneCountInString(content) > siteInfo.MaxPostLength
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
	num, _ := db.NextId(name)
	return toB64(num + 1)
}

func toB64(num int64) string {
	buf := make([]byte, 8)
	binary.PutVarint(buf, num)
	btr := bytes.TrimRight(buf, "\000")
	str := base64.URLEncoding.EncodeToString(btr)
	return strings.TrimRight(str, "=")
}

func fromB64(str string) (int64, error) {
	// Add padding
	str += strings.Repeat("=", 4-(len(str)%4))
	buf, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		return 0, err
	}
	num, _ := binary.Varint(buf)
	return num, nil
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
		return b64[:length]
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func strdate(unixtime int64) string {
	return time.Unix(unixtime, 0).Format("02/01/2006 15:04")
}

// getHiddenElems checks if the request contains a cookie 'crHidden'
// and parses its value, returning a slice with hidden threads and
// hidden tags.
func getHiddenElems(req *http.Request) (threads, tags []string) {
	if cookie, err := req.Cookie("crHidden"); err == nil {
		// cookie value has the format: "url1&url2&tag1&..."
		val, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			return
		}
		splitted := strings.Split(val, "&")
		for _, s := range splitted {
			if len(s) < 1 {
				continue
			}
			if s[0] == '#' {
				tags = append(tags, s[1:])
			} else {
				threads = append(threads, s)
			}
		}
	}
	return
}

// retreiveThreads retreives up to `n` threads, skipping the first `offset`,
// from the DB, excluding the ones matching the hidden elements of the client.
func retreiveThreads(n, offset int, hThreads, hTags []string) ([]ThreadInfo, error) {
	threads, err := db.GetThreadList("", n, offset, hThreads, hTags)
	if err != nil {
		return nil, err
	}

	tinfos := make([]ThreadInfo, 0, siteInfo.HomeThreadsNum)
	for i := range threads {
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
		if tinfos[i].Page < 1 {
			tinfos[i].Page = 1
		}
	}
	return tinfos, err
}

func randomCaptcha() (CaptchaData, error) {
	if len(captchas) < 1 {
		return CaptchaData{}, errCapthasNotConfigured
	}
	return captchas[mathrand.Intn(len(captchas))], nil
}

// html.EscapeString does NOT escape slash by itself; this function does.
func htmlFullEscape(str string) string {
	return strings.Replace(html.EscapeString(str), "/", "&sol;", -1)
}

func loadJson(path string, out interface{}) error {
	file, err := os.Open(maudRoot + string(os.PathSeparator) + path)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	return decoder.Decode(out)
}

// saveJson unmarshals the content of a struct in a JSON file, optionally
// backing up old file if existing.
func saveJson(in interface{}, path string, backup bool) error {
	if backup {
		if _, err := os.Stat(path); os.IsExist(err) {
			err := os.Rename(path, path+".old")
			if err != nil {
				return err
			}
		}
	}
	file, err := os.Create(maudRoot + string(os.PathSeparator) + path)
	var buf []byte
	out := bytes.NewBuffer(buf)
	enc, err := json.MarshalIndent(in, "", "\t")
	if err != nil {
		return err
	}
	_, err = out.Write(enc)
	if err != nil {
		return err
	}
	_, err = out.WriteTo(file)
	return err
}

// Emoji link data and generator

var emojis = []rune{'💯', '🔢', '👍', '👎', '🎱', '🅰', '🚡', '✈', '️', '👽', '🚑', '⚓', '️', '👼', '💢', '😠', '🐜', '🍎', '⬇', '️', '⬅', '️', '➡', '️', '⬆', '️', '🎨', '🏧', '👶', '🍼', '🎈', '🎍', '🍌', '‼', '️', '📊', '💈', '🏀', '🛀', '🔋', '🐻', '🐝', '🍺', '🐞', '🔔', '🚴', '👙', '🎂', '🃏', '🌼', '📘', '🚙', '💙', '🐗', '⛵', '️', '💣', '📖', '📚', '💥', '👢', '💐', '🎳', '👦', '🍞', '👰', '💼', '💔', '🐛', '💡', '🚌', '👤', '🌵', '🍰', '📆', '🐫', '🍬', '🚗', '🐈', '💿', '🏁', '🍒', '🐔', '🍫', '🎄', '👏', '🎬', '☁', '️', '🍸', '☕', '️', '💻', '🎊', '🚧', '🍪', '🌽', '👫', '🐄', '🐊', '👑', '🔮', '💘', '🌀', '💃', '🎯', '💨', '🌳', '💫', '🐕', '💵', '🐬', '🍩', '🚪', '💧', '👂', '🌍', '👓', '👊', '👣', '🔥', '🔦', '💾', '🏈', '🍀', '🍤', '🍟', '🐸', '🎲', '💎', '👻', '🎁', '🌐', '⛳', '️', '🍇', '🍏', '🎸', '🔫', '🍔', '🔨', '👜', '🐣', '🌿', '🔆', '👠', '🐴', '⌛', '️', '🍨', '🎃', '🔑', '💋', '🐨', '🍃', '🔗', '🔒', '📢', '🔍', '📫', '🍁', '📣', '📝', '🎤', '🔬', '💰', '🐒', '🌙', '🐁', '🎥', '💪', '🍄', '🎹', '🎵', '👔', '📰', '🔕', '⛔', '️', '🚫', '👃', '🔩', '🐙', '👌', '👐', '🐂', '🐼', '⛅', '️', '🐾', '🍑', '🍐', '🐧', '🎭', '☎', '️', '🐖', '🐽', '💊', '🍍', '🍕', '👇', '👈', '👉', '👆', '🚓', '🐩', '💩', '📯', '🍗', '🙏', '👛', '📌', '🐇', '🐎', '📻', '🐀', '🎀', '🍚', '💍', '🚀', '🐓', '🌹', '🚨', '📍', '🏃', '🎅', '📡', '🎷', '✂', '️', '🐚', '👕', '🚿', '💀', '😄', '🐌', '🐍', '❄', '️', '⛄', '️', '😭', '⚽', '️', '🔉', '👾', '💬', '⭐', '️', '🍓', '😎', '💦', '🏊', '💉', '🔭', '🎾', '💭', '🚽', '👅', '🎩', '🚥', '🏆', '🎺', '🐢', '🚦', '📼', '🎮', '🎻', '⌚', '️', '🐋', '🍷', '🐺', '🔧', '⚡', '️', '💤'}

func emojiLink(surl string) string {
	num, _ := fromB64(surl)
	buf := make([]byte, 8)
	binary.PutVarint(buf, num)
	btr := bytes.TrimRight(buf, "\000")
	str := ""
	for _, b := range btr {
		str += string(emojis[b])
	}
	return str
}
