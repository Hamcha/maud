package lightify

import (
	"../.."
	"regexp"
	"strings"
)

func Provide() *LightifyFormatter {
	lightify := new(LightifyFormatter)
	lightify.Init()
	return lightify
}

type LightifyFormatter struct {
	imgRgx    *regexp.Regexp
	derpiRgx  *regexp.Regexp
	iframeRgx *regexp.Regexp
}

func (f *LightifyFormatter) Init() {
	f.imgRgx = regexp.MustCompile(`(?:<a [^>]+>)?<img .*src=("[^"]+"|'[^']+'|[^'"][^\s]+).*>(?:</a>)?`)
	f.derpiRgx = regexp.MustCompile(`(?:img[0-9]\.)?derpicdn\.net`)
	f.iframeRgx = regexp.MustCompile(`<iframe .*src=("[^"]+"|'[^']+'|[^'"][^\s]+).*>`)
}

// Lightify takes a string containing HTML and converts embedded images and iframes
// to hyperlink to those resources. Images from Imgur or Derpibooru get converted
// into thumbnails linking to the full size resource.
func (f *LightifyFormatter) Format(content string) string {
	for _, match := range f.imgRgx.FindAllStringSubmatch(content, -1) {
		url := match[1]
		spl := strings.Split(url, "/")
		switch {
		case len(spl) > 2 && spl[2] == "i.imgur.com":
			content = strings.Replace(content, match[0], wrapImg(url, "<img src=\""+imgurThumb(url)+"\" alt="+url+"/>"), 1)
		case len(spl) > 2 && f.derpiRgx.MatchString(spl[2]):
			content = strings.Replace(content, match[0], wrapImg(url, "<img src=\""+derpibooruThumb(url)+"\" alt="+url+"/>"), 1)
		}
	}
	return content
}

// ReplaceTags replaces all <img> and <iframe> tags with hyperlinks to those resources.
func (f *LightifyFormatter) ReplaceTags(data modules.PostMutatorData) {
	content := *data.Content
	for _, match := range f.imgRgx.FindAllStringSubmatch(content, -1) {
		url := match[1]
		spl := strings.Split(url, "/")
		switch {
		case len(spl) > 2 && spl[2] == "i.imgur.com":
			continue
		case len(spl) > 2 && f.derpiRgx.MatchString(spl[2]):
			continue
		default:
			content = strings.Replace(content, match[0], "<a class='toggleImage' data-url="+url+">[Click to view image]</a>", 1)
		}
	}
	*data.Content = f.iframeRgx.ReplaceAllString(content, "<a target=\"_blank\" href=$1>[Click to open embedded content]</a>")
}

//// Unexported ////

func wrapImg(url, content string) string {
	return `<a target="_blank" href=` + url + ">" + content + "</a>"
}

func imgurThumb(origUrl string) string {
	/* origUrl must be like 'https://i.imgur.com/{id}.jpg', else the returned
	 * Url won't make sense. Getting a medium thumbnail just means
	 * inserting a 'm' before the image extension.
	 */
	idx := strings.LastIndex(origUrl, ".")
	thumb := origUrl[0:idx] + "m" + origUrl[idx:]
	return strings.Trim(thumb, `"'`)
}

func derpibooruThumb(origUrl string) string {
	splitted := strings.Split(origUrl, "/")
	if len(splitted) < 8 {
		return origUrl
	}
	/* Derpibooru's URLs are slightly more complex than Imgur ones.
	 * 5th element in the url is either 'view', which means a full size image,
	 * or something else, which means a thumbnail. In any case, we want
	 * the url to become https://img0.derpicdn.net/img/xxxx/yy/zz/{ID}/thumb.jpg,
	 * so we save xxxx, yy, zz and ID.
	 */
	fields := make([]string, 4)
	var ext string
	idx := strings.LastIndex(origUrl, "?")
	if idx < 0 {
		ext = origUrl[strings.LastIndex(origUrl, ".")+1:]
	} else {
		ext = origUrl[strings.LastIndex(origUrl, ".")+1 : idx]
	}
	i := 4
	var id string
	if splitted[4] == "view" {
		i++
		idx = strings.Index(splitted[8], "_")
		if idx < 0 {
			return ""
		}
		id = splitted[8][0:idx]
	} else {
		id = splitted[7]
	}
	copy(fields, splitted[i:i+3])
	thumb := strings.Join(splitted[0:4], "/") + "/" + strings.Join(fields, "/") + id + "/thumb." + ext
	return strings.Trim(thumb, `"'`)
}
