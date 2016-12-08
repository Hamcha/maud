package lightify

import (
	"log"
	"regexp"
	"strings"

	"github.com/hamcha/maud/maud/modules"
)

var (
	imgRgx    *regexp.Regexp
	derpiRgx  *regexp.Regexp
	iframeRgx *regexp.Regexp
	videoRgx  *regexp.Regexp
)

func init() {
	imgRgx = regexp.MustCompile(`(?:<a [^>]+>)?<img [^>]*src=("[^"]+"|'[^']+'|[^'"][^\s]+)[^>]*>(?:</a>)?`)
	derpiRgx = regexp.MustCompile(`(?:img[0-9]\.)?derpicdn\.net`)
	iframeRgx = regexp.MustCompile(`<iframe [^>]*src=("[^"]+"|'[^']+'|[^'"][^\s]+)[^>]*>`)
	videoRgx = regexp.MustCompile(`<video [^>]*src=("[^"]+"|'[^']+'|[^'"][^\s]+)[^>]*</video>`)
	log.Printf("[ OK ] Module initialized: Lightify")
}

type LightifyFormatter struct{}

func Provide() *LightifyFormatter {
	return &LightifyFormatter{}
}

// Format thumbnailifies all images from Imgur or Derpibooru
// and returns the other content unaltered
func (f *LightifyFormatter) Format(content string) string {
	for _, match := range imgRgx.FindAllStringSubmatch(content, -1) {
		url := match[1]
		spl := strings.Split(url, "/")
		switch {
		case len(spl) > 2 && strings.ToLower(spl[2]) == "i.imgur.com":
			content = strings.Replace(content, match[0], wrapImg(url, imgurThumb(url)), -1)
		case len(spl) > 2 && derpiRgx.MatchString(strings.ToLower(spl[2])):
			content = strings.Replace(content, match[0], wrapImg(
				url, "<img src='"+derpibooruThumb(url)+"' alt="+url+"/>"), 1)
		}
	}
	return content
}

// ReplaceTags replaces all <img> and <iframe> tags (except for thumbnails)
// with clickable links to get those resources. This is used when light mode
// is active, while Format is always used.
func (f *LightifyFormatter) ReplaceTags(data modules.PostMutatorData) {
	content := (*data.Post).Content
	for _, match := range imgRgx.FindAllStringSubmatch(content, -1) {
		url := match[1]
		spl := strings.Split(url, "/")
		switch {
		case len(spl) > 2 && strings.ToLower(spl[2]) == "i.imgur.com":
			continue
		case len(spl) > 2 && derpiRgx.MatchString(strings.ToLower(spl[2])):
			continue
		default:
			content = strings.Replace(content, match[0], "<a class='toggleImage' target='_blank' href="+url+">[Click to view image]</a>", 1)
		}
	}

	// Detect youtube embedded videos and give them a prettier link
	for _, match := range iframeRgx.FindAllStringSubmatch(content, -1) {
		url := match[1]
		spl := strings.Split(url, "/")
		switch {
		case len(spl) > 2 && strings.HasSuffix(strings.ToLower(spl[2]), "youtube.com"):
			full := strings.Replace(url, "/embed/", "/watch?v=", 1)
			content = strings.Replace(content, match[0], "<a target='_blank' href="+full+">[Youtube: "+full+"]</a><br />", 1)
		default:
			content = strings.Replace(content, match[0], "<a target='_blank' href="+url+">[Embedded: "+url+"]</a><br />", 1)
		}
	}

	for _, match := range videoRgx.FindAllStringSubmatch(content, -1) {
		url := strings.Trim(match[1], `"'`)
		if len(url) > 5 && url[len(url)-5:] == ".webm" {
			spl := strings.Split(url, "/")
			if len(spl) > 2 && spl[2] == "i.imgur.com" {
				gifv := url[:len(url)-5] + "m.gifv"
				content = strings.Replace(content, match[0], wrapImg(
					url, "<img src='"+gifv+"' alt="+url+"/>"), 1)
				continue
			}
		}
		content = strings.Replace(content, match[0], "<a target='_blank' href='"+url+"'>[Video: "+url+"]</a>", 1)
	}
	(*data.Post).Content = content
}

//// Unexported ////
func wrapImg(url, content string) string {
	return `<a target="_blank" href=` + url + ">" + content + "</a>"
}

func imgurThumb(origUrl string) string {
	/* origUrl must be like 'https://i.imgur.com/{id}.jpg', else the returned
	 * Url won't make sense. Getting a medium thumbnail just means
	 * inserting a 'm' before the image extension.
	 * If the image is a gif, though, we replace it with a webm instead of
	 * thumbnailifying it, since Imgur provides this opportunity.
	 * Note that if light mode is active, this will be converted again into
	 * a gifv thumbnail.
	 */
	url := strings.Trim(origUrl, `"'`)
	idx := strings.LastIndex(url, ".")

	ext := url[idx:]
	// Convert gif/gifvs to webm
	if ext == ".gif" || ext == ".gifv" {
		thumb := url[:idx] + ".webm"
		return `<video height="250" src="` + thumb +
			`" autoplay loop muted>[Your browser is unable to play this video]</video>`
	}

	/* If the image ends with a thumbnail suffix, it *may* be already a
	 * thumbnail. In this case, don't modify the url, or we may link an
	 * inexisting image by appending the thumbnail suffix 2 times.
	 */
	switch url[idx-1] {
	case 's', 'b', 't', 'm', 'l', 'h':
		return "<img src=\"" + url + "\" alt=\"" + url + "\"/>"
	}

	thumb := url[:idx] + "m" + ext
	return "<img src=\"" + thumb + "\" alt=\"" + url + "\"/>"
}

func derpibooruThumb(origUrl string) string {
	splitted := strings.Split(origUrl, "/")
	if len(splitted) < 8 {
		return strings.Trim(origUrl, `"'`)
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
		// Usually basename of the image is numericID_tags.ext;
		// sometimes it's just numericID.ext.
		idx = strings.Index(splitted[8], "_")
		if idx < 0 {
			idx = strings.LastIndex(splitted[8], ".")
			if idx < 0 {
				// Invalid url
				return ""
			}
		}
		id = splitted[8][:idx]
	} else {
		id = splitted[7]
	}
	copy(fields, splitted[i:i+3])
	thumb := strings.Join(splitted[:4], "/") + "/" + strings.Join(fields, "/") + id + "/thumb." + ext
	return strings.Trim(thumb, `"'`)
}
